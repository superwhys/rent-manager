package auth_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/miebyte/authkit"
	authmysql "github.com/miebyte/authkit/mysql"
	"github.com/miebyte/authkit/mysql/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"

	appcore "github.com/superwhys/rent-manager/internal/app"
	appauth "github.com/superwhys/rent-manager/internal/app/auth"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql"
	"github.com/superwhys/rent-manager/internal/xerr"
)

// sqliteAuthDialector keeps authkit's schema while adapting MySQL-only column types.
type sqliteAuthDialector struct{ sqlite.Dialector }

func (d sqliteAuthDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return sqlite.Migrator{Migrator: migrator.Migrator{Config: migrator.Config{
		DB: db, Dialector: d, CreateIndexAfterCreateTable: true,
	}}}
}

func (d sqliteAuthDialector) DataTypeOf(field *schema.Field) string {
	if strings.HasPrefix(string(field.DataType), "datetime") {
		return "datetime"
	}
	if strings.Contains(string(field.DataType), "CHARACTER SET") {
		return "text COLLATE BINARY"
	}
	return d.Dialector.DataTypeOf(field)
}

type codeSender struct{ code string }

func (s *codeSender) SendCode(_ context.Context, _, code string) error {
	s.code = code
	return nil
}

type authFixture struct {
	service *appauth.Service
	db      *gorm.DB
	sender  *codeSender
}

func newAuthFixture(t *testing.T) authFixture {
	t.Helper()
	db, err := gorm.Open(
		sqliteAuthDialector{
			sqlite.Dialector{DSN: fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
		},
		&gorm.Config{
			TranslateError: true, Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(authmysql.Models()...); err != nil {
		t.Fatal(err)
	}
	sender := &codeSender{}
	service, err := appauth.NewService(authmysql.Bind(db), mysql.NewAccountProfileRepository(db), sender, authkit.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return authFixture{service: service, db: db, sender: sender}
}

func (f authFixture) sendCode(t *testing.T, email string) string {
	t.Helper()
	if err := f.db.Model(&models.Challenge{}).
		Where("email = ?", strings.ToLower(strings.TrimSpace(email))).
		Update("sent", time.Now().Add(-2*time.Minute)).
		Error; err != nil {
		t.Fatal(err)
	}
	if err := f.service.SendCode(
		context.Background(),
		authkit.SendCodeInput{Email: email, IP: "192.0.2.8"},
	); err != nil {
		t.Fatal(err)
	}
	return f.sender.code
}

func (f authFixture) register(t *testing.T) *authkit.LoginResult {
	t.Helper()
	code := f.sendCode(t, "tenant@example.com")
	result, err := f.service.Register(context.Background(), appauth.RegisterRequest{
		Email: " TENANT@EXAMPLE.COM ", Code: code, Password: "  original-password  ",
	}, "192.0.2.8")
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRegisterCreatesBothBindingsAndValidSession(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()
	registered := f.register(t)
	if !registered.Created || registered.Account.Email != "tenant@example.com" {
		t.Fatalf("unexpected registration: %+v", registered.Account)
	}
	session, err := f.service.AuthenticateSession(ctx, registered.Token)
	if err != nil || session.Account.ID != registered.Account.ID {
		t.Fatalf("registration returned an invalid session: %v", err)
	}
	var count int64
	if err := f.db.Model(&models.Session{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("registration must leave exactly one session, count=%d, err=%v", count, err)
	}
	for _, method := range []string{authkit.MethodEmail, authkit.MethodPassword} {
		var binding models.Binding
		if err := f.db.Where("method = ? AND identifier = ?", method, registered.Account.Email).
			Take(&binding).
			Error; err != nil {
			t.Fatal(err)
		}
		if binding.AccountID == nil || *binding.AccountID != registered.Account.ID {
			t.Fatalf("%s binding does not belong to registered account", method)
		}
	}
	passwordLogin, err := f.service.LoginPassword(ctx, authkit.PasswordLoginInput{
		Identifier: " TENANT@EXAMPLE.COM ", Password: "  original-password  ", IP: "192.0.2.8",
	})
	if err != nil || passwordLogin.Account.ID != registered.Account.ID || passwordLogin.Created {
		t.Fatalf("password login did not reuse registered account: %v", err)
	}
	code := f.sendCode(t, registered.Account.Email)
	emailLogin, err := f.service.LoginEmail(
		ctx,
		authkit.EmailLoginInput{Email: registered.Account.Email, Code: code},
	)
	if err != nil || emailLogin.Account.ID != registered.Account.ID || emailLogin.Created {
		t.Fatalf("code login did not reuse registered account: %v", err)
	}
	if _, err := f.service.LoginEmail(
		ctx,
		authkit.EmailLoginInput{Email: registered.Account.Email, Code: code},
	); !errors.Is(
		err,
		authkit.ErrChallengeInvalid,
	) {
		t.Fatalf("consumed code remained usable: %v", err)
	}
}

func TestRegistrationCannotResetExistingPassword(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()
	registered := f.register(t)
	code := f.sendCode(t, registered.Account.Email)
	if _, err := f.service.Register(ctx, appauth.RegisterRequest{
		Email: registered.Account.Email, Code: code, Password: "replacement-password",
	}, "192.0.2.8"); !errors.Is(err, authkit.ErrConflict) {
		t.Fatalf("existing account registration = %v, want conflict", err)
	}
	if _, err := f.service.LoginPassword(ctx, authkit.PasswordLoginInput{
		Identifier: registered.Account.Email, Password: "  original-password  ", IP: "192.0.2.8",
	}); err != nil {
		t.Fatalf("existing password was changed: %v", err)
	}
	if _, err := f.service.LoginPassword(ctx, authkit.PasswordLoginInput{
		Identifier: registered.Account.Email, Password: "replacement-password", IP: "192.0.2.8",
	}); !errors.Is(err, authkit.ErrInvalidCredentials) {
		t.Fatalf("replacement password unexpectedly worked: %v", err)
	}
	if _, err := f.service.AuthenticateSession(ctx, registered.Token); err != nil {
		t.Fatalf("rejected registration revoked original session: %v", err)
	}
}

func TestLoginDoesNotRegisterUnknownEmail(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()
	code := f.sendCode(t, "unknown@example.com")
	if _, err := f.service.LoginEmail(
		ctx,
		authkit.EmailLoginInput{Email: "unknown@example.com", Code: code},
	); !errors.Is(
		err,
		authkit.ErrRegistrationDenied,
	) {
		t.Fatalf("unknown code login = %v", err)
	}
	if _, err := f.service.LoginPassword(
		ctx,
		authkit.PasswordLoginInput{Identifier: "unknown@example.com", Password: "some-password"},
	); !errors.Is(
		err,
		authkit.ErrInvalidCredentials,
	) {
		t.Fatalf("unknown password login = %v", err)
	}
	var count int64
	if err := f.db.Model(&models.Account{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("login created accounts: count=%d err=%v", count, err)
	}
}

func TestRegistrationRejectsInvalidPasswordsWithoutConsumingCode(t *testing.T) {
	for _, password := range []string{"short", strings.Repeat("x", 129)} {
		t.Run(fmt.Sprintf("length-%d", len(password)), func(t *testing.T) {
			f := newAuthFixture(t)
			code := f.sendCode(t, "tenant@example.com")
			_, err := f.service.Register(context.Background(), appauth.RegisterRequest{
				Email: "tenant@example.com", Code: code, Password: password,
			}, "192.0.2.8")
			if !errors.Is(err, authkit.ErrInvalidInput) {
				t.Fatalf("invalid password registration = %v", err)
			}
			var count int64
			if err := f.db.Model(&models.Account{}).Count(&count).Error; err != nil || count != 0 {
				t.Fatalf("invalid registration left account: count=%d err=%v", count, err)
			}
			if _, err := f.service.Register(context.Background(), appauth.RegisterRequest{
				Email: "tenant@example.com", Code: code, Password: "valid-password",
			}, "192.0.2.8"); err != nil {
				t.Fatalf("invalid password consumed code: %v", err)
			}
		})
	}
}

func TestRegistrationCommitsFailedCodeAttempts(t *testing.T) {
	f := newAuthFixture(t)
	code := f.sendCode(t, "tenant@example.com")
	wrongCode := "000000"
	if code == wrongCode {
		wrongCode = "111111"
	}
	for attempt := 1; attempt <= authkit.MaxAttempts; attempt++ {
		if _, err := f.service.Register(context.Background(), appauth.RegisterRequest{
			Email: "tenant@example.com", Code: wrongCode, Password: "valid-password",
		}, "192.0.2.8"); !errors.Is(err, authkit.ErrChallengeMismatch) {
			t.Fatalf("attempt %d = %v", attempt, err)
		}
		var challenge models.Challenge
		if err := f.db.Take(&challenge).Error; err != nil || challenge.Attempts != attempt {
			t.Fatalf("attempt was rolled back: attempts=%d err=%v", challenge.Attempts, err)
		}
	}
	if _, err := f.service.Register(context.Background(), appauth.RegisterRequest{
		Email: "tenant@example.com", Code: code, Password: "valid-password",
	}, "192.0.2.8"); !errors.Is(err, authkit.ErrChallengeInvalid) {
		t.Fatalf("exhausted code accepted: %v", err)
	}
}

func TestRegistrationRollsBackWhenFinalSessionFails(t *testing.T) {
	f := newAuthFixture(t)
	code := f.sendCode(t, "tenant@example.com")
	failure := errors.New("session storage failed")
	if err := f.db.Callback().
		Create().
		Before("gorm:create").
		Register("test:reject-password-session", func(tx *gorm.DB) {
			if session, ok := tx.Statement.Dest.(*models.Session); ok &&
				session.Method == authkit.MethodPassword {
				tx.AddError(failure)
			}
		}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Register(context.Background(), appauth.RegisterRequest{
		Email: "tenant@example.com", Code: code, Password: "valid-password",
	}, "192.0.2.8"); !errors.Is(err, failure) {
		t.Fatalf("registration did not return final session error: %v", err)
	}
	for _, model := range []any{&models.Account{}, &models.Session{}} {
		var count int64
		if err := f.db.Model(model).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("failed registration left %T records: count=%d err=%v", model, count, err)
		}
	}
	if err := f.db.Callback().Create().Remove("test:reject-password-session"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Register(context.Background(), appauth.RegisterRequest{
		Email: "tenant@example.com", Code: code, Password: "valid-password",
	}, "192.0.2.8"); err != nil {
		t.Fatalf("rolled back registration could not retry: %v", err)
	}
}

func TestRegistrationRateLimitDoesNotLeavePartialAccount(t *testing.T) {
	f := newAuthFixture(t)
	code := f.sendCode(t, "tenant@example.com")
	rate := models.Rate{
		ID:     fmt.Sprintf("%x", sha256.Sum256([]byte("password-ip:192.0.2.8"))),
		Starts: time.Now().UTC(), Hits: authkit.PasswordIPRateLimit,
	}
	if err := f.db.Create(&rate).Error; err != nil {
		t.Fatal(err)
	}
	request := appauth.RegisterRequest{
		Email:    "tenant@example.com",
		Code:     code,
		Password: "valid-password",
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := f.service.Register(
			context.Background(),
			request,
			"192.0.2.8",
		); !errors.Is(
			err,
			authkit.ErrTooManyRequests,
		) {
			t.Fatalf("limited registration = %v", err)
		}
		for _, model := range []any{&models.Account{}, &models.Session{}} {
			var count int64
			if err := f.db.Model(model).Count(&count).Error; err != nil || count != 0 {
				t.Fatalf("limited registration left %T records: count=%d err=%v", model, count, err)
			}
		}
		var stored models.Rate
		if err := f.db.Where("id = ?", rate.ID).
			Take(&stored).
			Error; err != nil ||
			stored.Hits != authkit.PasswordIPRateLimit {
			t.Fatalf("rollback reset IP rate: hits=%d err=%v", stored.Hits, err)
		}
	}
	if err := f.db.Model(&rate).Update("starts", time.Now().Add(-2*time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Register(context.Background(), request, "192.0.2.8"); err != nil {
		t.Fatalf("limited registration consumed code or left conflicting binding: %v", err)
	}
}

func TestUpdateAccountPersistsDisplayNameAndPreservesLogin(t *testing.T) {
	f := newAuthFixture(t)
	registered := f.register(t)
	ctx := appcore.SetAccountIDToContext(context.Background(), registered.Account.ID)
	other := authkit.Account{ID: strings.Repeat("b", 64), Username: "其他账号"}
	if err := authmysql.Bind(f.db).WithTransaction(ctx, func(repos authkit.Repositories) error {
		return repos.Accounts().Create(ctx, &other, authkit.Credential{
			Method: authkit.MethodEmail, Identifier: "other@example.com",
		})
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"房东 @团队", strings.Repeat("名", 64)} {
		account, err := f.service.UpdateAccount(ctx, appauth.UpdateAccountRequest{Username: "　 " + name + " 　"})
		if err != nil {
			t.Fatal(err)
		}
		if account.ID != registered.Account.ID || account.Username != name || account.Email != registered.Account.Email {
			t.Fatalf("updated account = %+v", account)
		}
		session, err := f.service.AuthenticateSession(ctx, registered.Token)
		if err != nil || session.Account != *account {
			t.Fatalf("session did not read saved account: %+v, %v", session, err)
		}
		if _, err := f.service.UpdateAccount(ctx, appauth.UpdateAccountRequest{Username: name}); err != nil {
			t.Fatalf("saving unchanged name failed: %v", err)
		}
	}
	storedOther, err := authmysql.Bind(f.db).Accounts().GetByID(ctx, other.ID)
	if err != nil || storedOther.Username != other.Username {
		t.Fatalf("another account changed: %+v, %v", storedOther, err)
	}
	login, err := f.service.LoginPassword(ctx, authkit.PasswordLoginInput{
		Identifier: registered.Account.Email, Password: "  original-password  ", IP: "192.0.2.8",
	})
	if err != nil || login.Account.ID != registered.Account.ID || login.Account.Username != strings.Repeat("名", 64) {
		t.Fatalf("display name update affected password login: %+v, %v", login, err)
	}
}

func TestUpdateAccountRejectsInvalidNamesWithoutChangingAccount(t *testing.T) {
	f := newAuthFixture(t)
	registered := f.register(t)
	ctx := appcore.SetAccountIDToContext(context.Background(), registered.Account.ID)
	for _, tt := range []struct {
		name     string
		username string
	}{
		{name: "empty"},
		{name: "whitespace", username: "　   "},
		{name: "long ASCII", username: strings.Repeat("a", 65)},
		{name: "long Unicode", username: strings.Repeat("名", 65)},
		{name: "embedded newline", username: "账号\n名称"},
		{name: "trailing newline", username: "账号\n"},
		{name: "tab", username: "\t账号"},
		{name: "NUL", username: "账号\x00"},
		{name: "DEL", username: "账号\x7f"},
		{name: "invalid UTF-8", username: "\xff"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := f.service.UpdateAccount(ctx, appauth.UpdateAccountRequest{Username: tt.username}); !errors.Is(err, authkit.ErrInvalidInput) {
				t.Fatalf("invalid display name = %v", err)
			}
			stored, err := authmysql.Bind(f.db).Accounts().GetByID(ctx, registered.Account.ID)
			if err != nil || *stored != registered.Account {
				t.Fatalf("rejected update changed account: %+v, %v", stored, err)
			}
		})
	}
}

func TestUpdateAccountRequiresIdentityAndPreservesRepositoryError(t *testing.T) {
	f := newAuthFixture(t)
	registered := f.register(t)
	request := appauth.UpdateAccountRequest{Username: "新账号名称"}
	if _, err := f.service.UpdateAccount(context.Background(), request); !errors.Is(err, xerr.ErrUnauthorized) {
		t.Fatalf("missing identity error = %v", err)
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), registered.Account.ID)
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := f.service.UpdateAccount(cancelled, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled request error = %v", err)
	}
	failure := errors.New("profile storage failed")
	if err := f.db.Callback().Update().Before("gorm:update").Register("test:reject-profile-update", func(tx *gorm.DB) {
		tx.AddError(failure)
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.UpdateAccount(ctx, request); !errors.Is(err, failure) {
		t.Fatalf("repository error was replaced: %v", err)
	}
	stored, err := authmysql.Bind(f.db).Accounts().GetByID(ctx, registered.Account.ID)
	if err != nil || *stored != registered.Account {
		t.Fatalf("failed update changed account: %+v, %v", stored, err)
	}
}
