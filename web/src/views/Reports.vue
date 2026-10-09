<script setup lang="ts">
import type { ReportData } from '@/types'
import { ref } from 'vue'

// 报表数据
const reportData = ref<ReportData>({
    incomeChart: [
        { label: '1月', value: 52000 },
        { label: '2月', value: 54000 },
        { label: '3月', value: 53500 },
        { label: '4月', value: 55000 },
        { label: '5月', value: 52800 },
        { label: '6月', value: 54200 },
        { label: '7月', value: 55500 }
    ],
    occupancyChart: [
        { label: '已出租', value: 85 },
        { label: '空置', value: 15 }
    ],
    utilityChart: [
        { label: '水费', value: 1200 },
        { label: '电费', value: 1800 },
        { label: '燃气费', value: 600 }
    ],
    yearComparison: {
        currentYear: 633600,
        lastYear: 598400,
        growthRate: '+5.9%'
    }
})

// 处理导出
const exportIncomeReport = () => {
    console.log('导出收入报表')
}

const exportOccupancyReport = () => {
    console.log('导出入住统计')
}

const exportUtilityReport = () => {
    console.log('导出水电报表')
}

const generateMonthlySummary = () => {
    console.log('生成月度总结')
}
</script>

<template>
    <div class="reports">
        <h2 class="section-title">📈 报表分析</h2>

        <div class="reports-grid">
            <ReportCard title="📊 收入趋势分析" :has-chart="true" />

            <ReportCard title="🏠 入住率统计" :has-chart="true" />

            <ReportCard title="💡 能耗分析" :has-chart="true" />

            <ReportCard title="📈 年度对比">
                <div class="summary-stats">
                    <div class="summary-item">
                        <span class="label">本年收入</span>
                        <span class="value">¥{{ reportData.yearComparison.currentYear.toLocaleString('zh-CN') }}</span>
                    </div>
                    <div class="summary-item">
                        <span class="label">去年收入</span>
                        <span class="value">¥{{ reportData.yearComparison.lastYear.toLocaleString('zh-CN') }}</span>
                    </div>
                    <div class="summary-item growth">
                        <span class="label">增长率</span>
                        <span class="value">{{ reportData.yearComparison.growthRate }}</span>
                    </div>
                </div>
            </ReportCard>
        </div>

        <div class="export-section">
            <h3>📤 报表导出</h3>
            <div class="export-buttons">
                <button class="btn btn-primary" @click="exportIncomeReport">
                    导出收入报表
                </button>
                <button class="btn btn-primary" @click="exportOccupancyReport">
                    导出入住统计
                </button>
                <button class="btn btn-primary" @click="exportUtilityReport">
                    导出水电报表
                </button>
                <button class="btn btn-success" @click="generateMonthlySummary">
                    生成月度总结
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.reports {
    padding: 20px;
}

.section-title {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 20px;
    color: #333;
}

.reports-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 25px;
    margin-bottom: 30px;
}

.summary-stats {
    display: flex;
    flex-direction: column;
    gap: 15px;
}

.summary-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid #f3f4f6;
}

.summary-item:last-child {
    border-bottom: none;
}

.summary-item.growth .value {
    color: #10b981;
    font-weight: bold;
}

.export-section {
    background: white;
    padding: 25px;
    border-radius: 12px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.export-section h3 {
    margin-bottom: 15px;
    color: #1f2937;
}

.export-buttons {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 15px;
    margin-top: 15px;
}

.btn {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-primary {
    background: #4f46e5;
    color: white;
}

.btn-primary:hover {
    background: #4338ca;
    transform: translateY(-2px);
}

.btn-success {
    background: #10b981;
    color: white;
}

.btn-success:hover {
    background: #059669;
    transform: translateY(-2px);
}

/* 响应式设计 */
@media (max-width: 480px) {
    .reports {
        padding: 10px;
    }
    
    .section-title {
        font-size: 20px;
        margin-bottom: 15px;
    }
    
    .reports-grid {
        grid-template-columns: 1fr;
        gap: 15px;
        margin-bottom: 20px;
    }
    
    .export-section {
        padding: 15px;
    }
    
    .export-section h3 {
        font-size: 16px;
        margin-bottom: 12px;
    }
    
    .export-buttons {
        grid-template-columns: 1fr;
        gap: 10px;
        margin-top: 12px;
    }
    
    .btn {
        padding: 10px 16px;
        font-size: 13px;
        width: 100%;
    }
    
    .summary-stats {
        gap: 10px;
    }
    
    .summary-item {
        padding: 8px 0;
        flex-wrap: wrap;
        gap: 5px;
    }
    
    .summary-item .label,
    .summary-item .value {
        font-size: 14px;
    }
}
</style>