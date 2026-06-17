package home

import (
	"fmt"
	"strconv"
)

// Summary 描述首页楼栋与房间汇总信息。
type Summary struct {
	BuildingCount               int
	RoomCount                   int
	RentedRoomCount             int
	VacantRoomCount             int
	CurrentMonth                string
	CurrentMonthIncomeCents     int
	PreviousMonthIncomeCents    int
	MonthlyIncomeGrowthCents    int
	CurrentMonthUnpaidBillCount int
	Todos                       []Todo
}

// Todo 描述首页待办项。
type Todo struct {
	Title       string
	Description string
	Tag         string
}

// BuildingRoomSummaryResponse 是首页楼栋与房间汇总响应。
type BuildingRoomSummaryResponse struct {
	BuildingCount   int `json:"building_count"`
	RoomCount       int `json:"room_count"`
	RentedRoomCount int `json:"rented_room_count"`
	VacantRoomCount int `json:"vacant_room_count"`
}

// StatResponse 是首页核心统计项响应。
type StatResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

// TodoResponse 是首页待办项响应。
type TodoResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
}

// Response 是首页聚合数据响应。
type Response struct {
	BuildingRoomSummary BuildingRoomSummaryResponse `json:"building_room_summary"`
	IncomeAmount        string                      `json:"income_amount"`
	IncomeDecimal       string                      `json:"income_decimal"`
	MonthlyGrowth       string                      `json:"monthly_growth"`
	CurrentMonth        string                      `json:"current_month"`
	Stats               []StatResponse              `json:"stats"`
	Todos               []TodoResponse              `json:"todos"`
}

// ToSummaryResponse 将首页服务汇总转换为接口响应。
func ToSummaryResponse(summary *Summary) *Response {
	if summary == nil {
		return &Response{}
	}
	incomeAmount, incomeDecimal := formatCentsParts(summary.CurrentMonthIncomeCents)
	return &Response{
		BuildingRoomSummary: BuildingRoomSummaryResponse{
			BuildingCount:   summary.BuildingCount,
			RoomCount:       summary.RoomCount,
			RentedRoomCount: summary.RentedRoomCount,
			VacantRoomCount: summary.VacantRoomCount,
		},
		IncomeAmount:  incomeAmount,
		IncomeDecimal: incomeDecimal,
		MonthlyGrowth: formatSignedCents(summary.MonthlyIncomeGrowthCents),
		CurrentMonth:  summary.CurrentMonth,
		// TODO:
		Stats: []StatResponse{
			{
				Value: strconv.Itoa(summary.BuildingCount),
				Label: "楼栋",
				Tone:  "green",
			},
			{
				Value: strconv.Itoa(summary.RoomCount),
				Label: "房间",
				Tone:  "blue",
			},
			{
				Value: strconv.Itoa(summary.CurrentMonthUnpaidBillCount),
				Label: "待收",
				Tone:  "orange",
			},
		},
		Todos: toTodoResponses(summary.Todos),
	}
}

func toTodoResponses(todos []Todo) []TodoResponse {
	responses := make([]TodoResponse, 0, len(todos))
	for _, todo := range todos {
		responses = append(responses, TodoResponse{
			Title:       todo.Title,
			Description: todo.Description,
			Tag:         todo.Tag,
		})
	}
	return responses
}

// formatCentsParts 将分金额拆成整数金额和小数金额字符串。
func formatCentsParts(cents int) (string, string) {
	value := int64(cents)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	return sign + formatIntegerWithComma(value/100), fmt.Sprintf(".%02d", value%100)
}

// formatSignedCents 将分金额格式化为带人民币符号的增减金额。
func formatSignedCents(cents int) string {
	value := int64(cents)
	sign := "+"
	if value < 0 {
		sign = "-"
		value = -value
	}
	amount, decimal := formatCentsParts(int(value))
	if decimal == ".00" {
		return sign + "¥" + amount
	}
	return sign + "¥" + amount + decimal
}

// formatIntegerWithComma 将整数格式化为带千分位的字符串。
func formatIntegerWithComma(value int64) string {
	text := strconv.FormatInt(value, 10)
	if len(text) <= 3 {
		return text
	}
	firstGroupLength := len(text) % 3
	if firstGroupLength == 0 {
		firstGroupLength = 3
	}
	result := text[:firstGroupLength]
	for i := firstGroupLength; i < len(text); i += 3 {
		result += "," + text[i:i+3]
	}
	return result
}
