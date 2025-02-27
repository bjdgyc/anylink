<template>
    <div id="lock-manager">
        <el-card>
            <div slot="header">
                <el-button type="primary" @click="getLocks">刷新信息</el-button>
            </div>
            <el-table :data="locksInfo" style="width: 100%" border>
                <el-table-column type="index" label="序号" width="60"></el-table-column>
                <el-table-column prop="description" label="描述"></el-table-column>
                <el-table-column prop="username" label="用户名"></el-table-column>
                <el-table-column prop="ip" label="IP地址"></el-table-column>
                <el-table-column prop="state.locked" label="状态" width="100">
                    <template slot-scope="scope">
                        <el-tag :type="scope.row.state.locked ? 'danger' : 'success'">
                            {{ scope.row.state.locked ? '已锁定' : '未锁定' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="state.attempts" label="失败次数"></el-table-column>
                <el-table-column prop="state.lock_time" label="锁定截止时间">
                    <template slot-scope="scope">
                        {{ formatDate(scope.row.state.lock_time) }}
                    </template>
                </el-table-column>
                <el-table-column prop="state.lastAttempt" label="最后尝试时间">
                    <template slot-scope="scope">
                        {{ formatDate(scope.row.state.lastAttempt) }}
                    </template>
                </el-table-column>
                <el-table-column label="操作">
                    <template slot-scope="scope">
                        <div class="button">
                            <el-button size="small" type="danger" @click="unlock(scope.row)">
                                解锁
                            </el-button>
                        </div>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>
    </div>
</template>

<script>
import axios from 'axios';

export default {
    name: 'LockManager',
    data() {
        return {
            locksInfo: []
        };
    },
    methods: {
        getLocks() {
            axios.get('/locksinfo/list')
                .then(response => {
                    this.locksInfo = response.data.data;
                })
                .catch(error => {
                    console.error('Failed to get locks info:', error);
                    this.$message.error('无法获取锁信息，请稍后再试。');
                });
        },
        unlock(lock) {
            const lockInfo = {
                state: { locked: false },
                username: lock.username,
                ip: lock.ip,
                description: lock.description
            };

            axios.post('/locksinfo/unlok', lockInfo)
                .then(() => {
                    this.$message.success('解锁成功！');
                    this.getLocks();
                })
                .catch(error => {
                    console.error('Failed to unlock:', error);
                    this.$message.error('解锁失败，请稍后再试。');
                });
        },
        formatDate(dateString) {
            if (!dateString) return '';
            const date = new Date(dateString);
            return new Intl.DateTimeFormat('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
                hour12: false
            }).format(date);
        }
    },
    created() {
        this.getLocks();
    }
};
</script>

<style scoped>
.button {
    display: flex;
    justify-content: center;
    align-items: center;
}
</style>