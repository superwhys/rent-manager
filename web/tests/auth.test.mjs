import assert from 'node:assert/strict'
import test from 'node:test'
import {
  accountNameError,
  emailError,
  verificationCodeError,
  registrationPasswordError,
  loginPasswordError,
  safeAuthRedirect,
} from '../src/utils/authValidation.ts'

test('账号名称去除首尾空白，按 Unicode 字符限制长度并拒绝控制字符', () => {
  for (const name of ['  我的账号  ', '租序 @团队', '😀'.repeat(64)]) {
    assert.equal(accountNameError(name), '')
  }
  for (const name of ['', '   ', '名'.repeat(65), '😀'.repeat(65)]) {
    assert.ok(accountNameError(name))
  }
  for (const name of ['名称\n', '\t名称', '名\u0000称', '名\u0085称']) {
    assert.match(accountNameError(name), /控制字符/)
  }
})

test('邮箱校验接受首尾空白，拒绝空值、不完整地址与超长地址', () => {
  assert.equal(emailError('  Owner@example.com  '), '')
  for (const email of [
    '',
    ' ',
    'owner',
    'owner@',
    '@example.com',
    'a b@example.com',
    'a'.repeat(250) + '@example.com',
  ]) {
    assert.ok(emailError(email))
  }
})

test('验证码必须恰好为六位数字', () => {
  assert.equal(verificationCodeError('012345'), '')
  for (const code of [
    '',
    '12345',
    '1234567',
    '12345a',
    '１２３４５６',
    ' 123456',
  ]) {
    assert.ok(verificationCodeError(code))
  }
})

test('注册密码按 Unicode 字符计数，并原样比较确认密码', () => {
  assert.equal(
    registrationPasswordError('密😀'.repeat(4), '密😀'.repeat(4)),
    '',
  )
  assert.equal(
    registrationPasswordError('😀'.repeat(128), '😀'.repeat(128)),
    '',
  )
  assert.equal(registrationPasswordError(' pass12 ', ' pass12 '), '')
  assert.ok(registrationPasswordError('😀'.repeat(7), '😀'.repeat(7)))
  assert.ok(registrationPasswordError('😀'.repeat(129), '😀'.repeat(129)))
  assert.match(registrationPasswordError('password', ' password'), /不一致/)
})

test('密码登录允许已有短密码，不擅自清除空格', () => {
  assert.equal(loginPasswordError('short'), '')
  assert.equal(loginPasswordError(' '), '')
  assert.ok(loginPasswordError(''))
})

test('登录后只允许站内业务跳转，防止回到认证页面或外部地址', () => {
  assert.equal(
    safeAuthRedirect('/rooms?unit_id=2#list'),
    '/rooms?unit_id=2#list',
  )
  for (const path of [
    null,
    undefined,
    [],
    '//evil.example',
    'https://evil.example',
    '/\\evil.example',
    '/login',
    '/login?redirect=/rooms',
    '/register#form',
    '/%6Cogin',
    '/%2f%2fevil.example',
    '/rooms\n',
  ]) {
    assert.equal(safeAuthRedirect(path), '/dashboard')
  }
})
