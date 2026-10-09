# 🏠 智能房租管理系统

一个基于 Vue 3 + TypeScript + Vite 构建的现代化房租管理平台，为物业管理提供全方位的数字化解决方案。

![系统版本](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Vue版本](https://img.shields.io/badge/Vue-3.x-green.svg)
![TypeScript](https://img.shields.io/badge/TypeScript-5.x-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## 📋 项目概述

智能房租管理系统是一个专为房东、物业管理公司设计的综合性管理平台。系统提供直观的界面和强大的功能，帮助用户轻松管理房间、租客、账单、水电费用等各项业务。

### 🎯 核心特性

- 📊 **数据可视化** - 直观的仪表板展示关键指标
- 🏠 **房间管理** - 完整的房间信息维护和状态跟踪
- 👥 **租客管理** - 租客档案管理和合同状态跟踪
- 💰 **账单管理** - 自动化账单生成和收费统计
- ⚡ **水电管理** - 智能抄表记录和费用计算
- 📈 **报表分析** - 多维度数据分析和趋势展示
- ⚙️ **系统设置** - 灵活的系统配置和数据管理

## 🚀 技术栈

### 前端框架

- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全的 JavaScript 超集
- **Vite** - 下一代前端构建工具

### UI 组件

- **Element Plus** - Vue 3 企业级 UI 组件库
- **自定义组件** - 高度定制化的业务组件

### 开发工具

- **Vue Router** - 官方路由管理器
- **Pinia** - Vue 状态管理库（准备集成）
- **ESLint + Prettier** - 代码质量和格式化工具

## 📦 功能模块

### 1. 📊 仪表板 (Dashboard)

- **当月概览统计**
  - 总房间数统计
  - 已出租房间数量
  - 本月总收入
  - 本月总支出
- **房间状态总览**
  - 实时房间状态展示
  - 快速房间信息查看
  - 房间卡片交互

### 2. 🏠 房间管理 (Rooms)

- **房间信息管理**
  - 房间基本信息维护
  - 房间状态（已出租/空置/待缴费）
  - 租金设置和调整
- **数据展示**
  - 表格形式展示房间列表
  - 支持编辑、查看、删除操作
  - 按楼层和状态筛选

### 3. 👥 租客管理 (Tenants)

- **租客档案管理**
  - 个人信息登记
  - 联系方式维护
  - 身份证信息记录
- **合同管理**
  - 入住时间记录
  - 押金管理
  - 合同状态跟踪

### 4. 💰 账单管理 (Bills)

- **账单统计面板**
  - 本月应收金额
  - 已收金额统计
  - 未收金额提醒
  - 收费率计算
- **账单详情管理**
  - 账单生成和编辑
  - 费用明细查看
  - 支付状态跟踪
  - 批量导出功能

### 5. ⚡ 水电管理 (Utilities)

- **费用设置**
  - 水费单价和保底设置
  - 电费单价和保底设置
  - 燃气费单价和保底设置
- **抄表记录**
  - 月度抄表数据录入
  - 用量自动计算
  - 历史记录查询

### 6. 📈 报表分析 (Reports)

- **数据可视化**
  - 收入趋势分析图表
  - 入住率统计图表
  - 能耗分析图表
- **年度对比**
  - 年度收入对比
  - 增长率计算
  - 关键指标汇总
- **报表导出**
  - 多格式报表导出
  - 月度总结生成

### 7. ⚙️ 系统设置 (Settings)

- **基础信息设置**
  - 物业基本信息配置
  - 联系方式设置
- **费用设置**
  - 默认租金配置
  - 押金倍数设置
  - 附加费用设置
- **系统配置**
  - 账单生成规则
  - 提醒设置
  - 自动备份配置
- **安全设置**
  - 密码管理
  - 会话超时设置
- **数据管理**
  - 数据备份恢复
  - 过期数据清理

## 🛠️ 安装和使用

### 环境要求

- Node.js >= 16.0.0
- npm >= 8.0.0 或 yarn >= 1.22.0
- 现代浏览器（Chrome 88+, Firefox 85+, Safari 14+）

### 安装步骤

1. **克隆项目**

```bash
git clone <repository-url>
cd rent-manager-web
```

2. **安装依赖**

```bash
npm install
# 或
yarn install
```

3. **启动开发服务器**

```bash
npm run dev
# 或
yarn dev
```

4. **构建生产版本**

```bash
npm run build
# 或
yarn build
```

5. **预览生产构建**

```bash
npm run preview
# 或
yarn preview
```

## 📂 项目结构

```
rent-manager-web/
├── public/                 # 静态资源
├── src/
│   ├── components/         # 通用组件
│   │   ├── DashboardCard.vue    # 仪表板卡片
│   │   ├── DataTable.vue        # 数据表格
│   │   ├── NavItem.vue          # 导航项
│   │   ├── ReportCard.vue       # 报表卡片
│   │   ├── RoomCard.vue         # 房间卡片
│   │   ├── SettingsCard.vue     # 设置卡片
│   │   ├── StatCard.vue         # 统计卡片
│   │   ├── StatusBadge.vue      # 状态标签
│   │   └── UtilityCard.vue      # 水电设置卡片
│   ├── router/             # 路由配置
│   │   └── index.ts
│   ├── types/              # TypeScript 类型定义
│   │   └── index.ts
│   ├── views/              # 页面组件
│   │   ├── Dashboard.vue        # 仪表板
│   │   ├── Rooms.vue           # 房间管理
│   │   ├── Tenants.vue         # 租客管理
│   │   ├── Bills.vue           # 账单管理
│   │   ├── Utilities.vue       # 水电管理
│   │   ├── Reports.vue         # 报表分析
│   │   └── Settings.vue        # 系统设置
│   ├── App.vue             # 根组件
│   ├── main.ts             # 应用入口
│   └── style.css           # 全局样式
├── index.html              # HTML 模板
├── package.json            # 项目配置
├── README.md              # 项目文档
├── tsconfig.json          # TypeScript 配置
├── vite.config.ts         # Vite 配置
└── yarn.lock              # 依赖锁定文件
```

## 🎨 组件设计

### 通用组件

- **DataTable**: 可配置的数据表格组件，支持自定义列和操作
- **StatusBadge**: 状态标签组件，支持多种状态样式
- **StatCard**: 统计数据展示卡片
- **UtilityCard**: 水电费用设置卡片
- **SettingsCard**: 系统设置卡片容器
- **ReportCard**: 报表数据展示卡片

### 业务组件

- **DashboardCard**: 仪表板统计卡片
- **RoomCard**: 房间信息展示卡片
- **NavItem**: 侧边栏导航项

## 🔧 开发指南

### 代码规范

- 使用 TypeScript 进行类型定义
- 组件采用 Composition API + `<script setup>` 语法
- 遵循 Vue 3 最佳实践
- 使用中文注释，英文变量名

### 样式规范

- 使用 scoped CSS 避免样式污染
- 采用 BEM 命名规范
- 响应式设计，支持移动端适配
- 使用 CSS Grid 和 Flexbox 布局

### 数据流

- Props down, Events up 原则
- 使用 ref/reactive 管理组件状态
- 预留 Pinia 状态管理集成

## 🚧 待实现功能

- [ ] 用户认证和权限管理
- [ ] 数据持久化（后端 API 集成）
- [ ] 图表库集成（Chart.js/ECharts）
- [ ] 文件上传和导出功能
- [ ] 消息通知系统
- [ ] 移动端优化
- [ ] 国际化支持
- [ ] 主题切换功能

## 📄 License

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 邮箱：<admin@example.com>
- 电话：138-0000-0000

---

*智能房租管理系统 - 让房屋管理更简单、更高效！*
