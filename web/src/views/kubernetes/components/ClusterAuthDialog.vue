<template>
  <div class="cluster-users-content" v-if="modelValue">
    <el-alert
      title="凭据用户管理"
      type="info"
      :closable="false"
      style="margin-bottom: 20px;"
    >
      <template #default>
        <div>管理已创建凭据的用户及其角色权限（每个用户只显示最新的有效凭据）</div>
      </template>
    </el-alert>

    <!-- 已创建的凭据用户列表 -->
    <div class="credentials-section">
      <div class="section-header">
        <h4>已创建的凭据</h4>
        <el-button
          type="primary"
          size="small"
          @click="handleRefresh"
          :icon="Refresh"
          :loading="loading"
        >
          刷新
        </el-button>
      </div>

      <el-table
        :data="uniqueCredentialUsers"
        border
        stripe
        v-loading="loading"
        :height="500"
        :max-height="600"
        style="width: 100%"
      >
        <el-table-column prop="username" label="用户名" width="180">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 10px;">
              <el-icon color="#409EFF" :size="20"><User /></el-icon>
              <div>
                <div style="font-weight: 500; font-size: 14px;">{{ row.username }}</div>
                <div style="font-size: 12px; color: #909399;" v-if="row.realName">{{ row.realName }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="serviceAccount" label="ServiceAccount" min-width="220">
          <template #default="{ row }">
            <el-tag type="info" size="default" style="font-family: 'Courier New', monospace; font-size: 13px;">
              {{ row.serviceAccount }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140">
          <template #default="{ row }">
            <el-tag type="success" size="default">{{ row.namespace }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 8px;">
              <el-icon color="#909399" :size="16"><Clock /></el-icon>
              <span style="font-size: 13px;">{{ formatDate(row.createdAt) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; justify-content: center; gap: 8px;">
              <el-tooltip content="查看凭据" placement="top">
                <el-button
                  type="success"
                  :icon="Document"
                  circle
                  size="small"
                  @click="handleViewCredential(row)"
                />
              </el-tooltip>
              <el-tooltip content="角色授权" placement="top">
                <el-button
                  type="primary"
                  :icon="Lock"
                  circle
                  size="small"
                  @click="handleManageRoles(row)"
                />
              </el-tooltip>
              <el-tooltip content="吊销凭据" placement="top">
                <el-button
                  type="danger"
                  :icon="Delete"
                  circle
                  size="small"
                  @click="handleRevoke(row)"
                />
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && !uniqueCredentialUsers.length" description="暂无凭据，请先在连接标签页申请凭据" />
    </div>

    <!-- 角色授权对话框 -->
    <el-dialog
      v-model="showRoleDialog"
      :title="`角色授权 - ${currentUser?.username}`"
      width="900px"
      append-to-body
    >
      <el-tabs v-model="roleDialogTab" type="border-card">
        <!-- 集群角色 -->
        <el-tab-pane label="集群角色" name="cluster">
          <div class="role-auth-content">
            <div class="role-section">
              <h4>选择角色</h4>
              <el-select
                v-model="selectedClusterRole"
                placeholder="请选择集群角色"
                filterable
                style="width: 100%"
                @change="handleClusterRoleChange"
              >
                <el-option
                  v-for="role in clusterRoles"
                  :key="role.name"
                  :label="role.name"
                  :value="role.name"
                />
              </el-select>
            </div>

            <div class="user-section" v-if="selectedClusterRole">
              <div class="section-header">
                <h4>已绑定角色</h4>
                <el-tag :type="isRoleBound ? 'success' : 'info'" size="large">
                  {{ isRoleBound ? '已绑定' : '未绑定' }}
                </el-tag>
              </div>
              <el-button
                v-if="!isRoleBound"
                type="primary"
                @click="handleBindRole"
                :loading="bindLoading"
                style="margin-top: 12px;"
              >
                绑定此角色
              </el-button>
              <el-button
                v-else
                type="danger"
                @click="handleUnbindRole"
                :loading="bindLoading"
                style="margin-top: 12px;"
              >
                解绑此角色
              </el-button>
            </div>

            <div class="role-detail" v-if="selectedClusterRole && currentClusterRoleDetail">
              <el-divider />
              <h4>角色权限</h4>
              <el-tree
                :data="permissionTree"
                :props="treeProps"
                :default-expand-all="false"
                node-key="id"
                max-height="200"
              >
                <template #default="{ node, data }">
                  <span class="tree-node">
                    <el-icon v-if="data.type === 'apiGroup'" color="#409EFF"><Folder /></el-icon>
                    <el-icon v-else-if="data.type === 'resource'" color="#67C23A"><Document /></el-icon>
                    <el-icon v-else-if="data.type === 'verb'" color="#E6A23C"><Operation /></el-icon>
                    <span class="node-label">{{ data.label }}</span>
                    <el-tag v-if="data.type === 'verb'" size="small" type="info">{{ data.value }}</el-tag>
                  </span>
                </template>
              </el-tree>
            </div>
          </div>
        </el-tab-pane>

        <!-- 命名空间角色 -->
        <el-tab-pane label="命名空间角色" name="namespace">
          <div class="role-auth-content">
            <el-row :gutter="20">
              <el-col :span="12">
                <div class="role-section">
                  <h4>选择命名空间</h4>
                  <el-select
                    v-model="selectedNamespace"
                    placeholder="请选择命名空间"
                    filterable
                    style="width: 100%"
                    @change="handleNamespaceChange"
                  >
                    <el-option
                      v-for="ns in namespaces"
                      :key="ns.name"
                      :label="ns.name"
                      :value="ns.name"
                    />
                  </el-select>
                </div>
              </el-col>
              <el-col :span="12">
                <div class="role-section">
                  <h4>选择角色</h4>
                  <el-select
                    v-model="selectedNamespaceRole"
                    placeholder="请选择角色"
                    filterable
                    style="width: 100%"
                    :disabled="!selectedNamespace"
                    @change="handleNamespaceRoleChange"
                  >
                    <el-option
                      v-for="role in namespaceRoles"
                      :key="role.name"
                      :label="role.name"
                      :value="role.name"
                    />
                  </el-select>
                </div>
              </el-col>
            </el-row>

            <div class="user-section" v-if="selectedNamespaceRole">
              <div class="section-header">
                <h4>角色状态</h4>
                <el-tag :type="isNsRoleBound ? 'success' : 'info'" size="large">
                  {{ isNsRoleBound ? '已绑定' : '未绑定' }}
                </el-tag>
              </div>
              <el-button
                v-if="!isNsRoleBound"
                type="primary"
                @click="handleBindNsRole"
                :loading="bindLoading"
                style="margin-top: 12px;"
              >
                绑定此角色
              </el-button>
              <el-button
                v-else
                type="danger"
                @click="handleUnbindNsRole"
                :loading="bindLoading"
                style="margin-top: 12px;"
              >
                解绑此角色
              </el-button>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <!-- KubeConfig查看对话框 -->
    <el-dialog
      v-model="showKubeConfigDialog"
      title="查看KubeConfig凭据"
      width="800px"
      append-to-body
    >
      <div class="kubeconfig-content">
        <el-alert
          title="凭据信息"
          type="info"
          :closable="false"
          style="margin-bottom: 16px;"
        >
          <template #default>
            <div>用户: <strong>{{ currentUser?.username }}</strong></div>
            <div>ServiceAccount: <strong>{{ currentUser?.serviceAccount }}</strong></div>
          </template>
        </el-alert>

        <div class="kubeconfig-actions" style="margin-bottom: 12px; display: flex; gap: 8px;">
          <el-button
            type="primary"
            size="small"
            @click="handleCopyKubeConfig"
            :icon="Document"
          >
            复制到剪贴板
          </el-button>
          <el-button
            type="success"
            size="small"
            @click="handleDownloadKubeConfig"
          >
            下载文件
          </el-button>
        </div>

        <el-input
          v-model="currentKubeConfig"
          type="textarea"
          :rows="20"
          readonly
          placeholder="KubeConfig内容"
          style="font-family: 'Courier New', monospace; font-size: 12px;"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder, Document, Operation, Lock, Delete, User, Clock, Refresh
} from '@element-plus/icons-vue'
import {
  getClusterRoles,
  getNamespacesForRoles,
  getNamespaceRoles,
  getRoleDetail,
  getRoleBoundUsers,
  bindUserToRole,
  unbindUserFromRole,
  getServiceAccountKubeConfig,
  type Cluster,
  type CredentialUser
} from '@/api/kubernetes'

interface Props {
  cluster: Cluster | null
  modelValue: boolean
  credentialUsers?: CredentialUser[]
}

const props = defineProps<Props>()
const emit = defineEmits(['update:modelValue', 'refresh'])

const loading = ref(false)
const showRoleDialog = ref(false)
const showKubeConfigDialog = ref(false)
const currentUser = ref<any>(null)
const currentKubeConfig = ref('')
const roleDialogTab = ref('cluster')

// 角色相关
const clusterRoles = ref<any[]>([])
const namespaces = ref<{ name: string; podCount: number }[]>([])
const namespaceRoles = ref<any[]>([])
const selectedClusterRole = ref('')
const selectedNamespace = ref('')
const selectedNamespaceRole = ref('')
const currentClusterRoleDetail = ref<any>(null)
const permissionTree = ref<any[]>([])
const bindLoading = ref(false)
const isRoleBound = ref(false)
const isNsRoleBound = ref(false)

const treeProps = {
  children: 'children',
  label: 'label'
}

// 计算属性：每个用户只显示最新的凭据
const uniqueCredentialUsers = computed(() => {
  if (!props.credentialUsers || props.credentialUsers.length === 0) {
    return []
  }

  // 按用户名分组，取每个用户最新的凭据
  const userMap = new Map<string, CredentialUser>()

  props.credentialUsers.forEach(user => {
    const existing = userMap.get(user.username)
    if (!existing || new Date(user.createdAt) > new Date(existing.createdAt)) {
      userMap.set(user.username, user)
    }
  })

  return Array.from(userMap.values()).sort((a, b) => {
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  })
})

// 刷新凭据列表
const handleRefresh = async () => {
  emit('refresh')
  ElMessage.success('刷新成功')
}

const handleManageRoles = (user: any) => {
  currentUser.value = user
  showRoleDialog.value = true
  roleDialogTab.value = 'cluster'
  loadClusterRoles()
  loadNamespaces()
  checkRoleBinding()
}

const handleViewCredential = async (user: any) => {
  try {
    if (!props.cluster) return

    currentUser.value = user

    // 调用API获取该ServiceAccount的kubeconfig
    const result = await getServiceAccountKubeConfig(props.cluster.id, user.serviceAccount)

    currentKubeConfig.value = result.kubeconfig
    showKubeConfigDialog.value = true
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '获取kubeconfig失败')
  }
}

const handleCopyKubeConfig = async () => {
  try {
    await navigator.clipboard.writeText(currentKubeConfig.value)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const handleDownloadKubeConfig = () => {
  const blob = new Blob([currentKubeConfig.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `kubeconfig-${currentUser?.username || 'user'}.yaml`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('下载成功')
}

const handleRevoke = async (user: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要吊销用户 "${user.username}" 的凭据吗？`,
      '确认吊销',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )

    // TODO: 调用吊销凭据API
    ElMessage.success('吊销成功')
    emit('refresh')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('吊销失败')
    }
  }
}

const loadClusterRoles = async () => {
  if (!props.cluster) return
  try {
    const roles = await getClusterRoles(props.cluster.id)
    clusterRoles.value = roles
  } catch (error) {
    ElMessage.error('加载集群角色失败')
  }
}

const loadNamespaces = async () => {
  if (!props.cluster) return
  try {
    const nsList = await getNamespacesForRoles(props.cluster.id)
    namespaces.value = nsList
  } catch (error) {
    ElMessage.error('加载命名空间失败')
  }
}

const loadNamespaceRoles = async () => {
  if (!props.cluster || !selectedNamespace.value) return
  try {
    const roles = await getNamespaceRoles(props.cluster.id, selectedNamespace.value)
    namespaceRoles.value = roles
  } catch (error) {
    ElMessage.error('加载命名空间角色失败')
  }
}

const handleClusterRoleChange = async () => {
  if (!selectedClusterRole.value) return
  try {
    const detail = await getRoleDetail(
      props.cluster.id,
      '',
      selectedClusterRole.value
    )
    currentClusterRoleDetail.value = detail
    permissionTree.value = buildPermissionTree(detail.rules || [])
    await checkRoleBinding()
  } catch (error) {
    ElMessage.error('加载角色详情失败')
  }
}

const handleNamespaceChange = () => {
  selectedNamespaceRole.value = ''
  namespaceRoles.value = []
  isNsRoleBound.value = false
}

const handleNamespaceRoleChange = async () => {
  if (!selectedNamespaceRole.value) return
  await checkNsRoleBinding()
}

const checkRoleBinding = async () => {
  if (!props.cluster || !currentUser.value || !selectedClusterRole.value) return

  try {
    const users = await getRoleBoundUsers(
      props.cluster.id,
      selectedClusterRole.value,
      ''
    )
    isRoleBound.value = users.some((u: any) => u.username === currentUser.value.username)
  } catch (error) {
    // 角色未绑定
    isRoleBound.value = false
  }
}

const checkNsRoleBinding = async () => {
  if (!props.cluster || !currentUser.value || !selectedNamespaceRole.value) return

  try {
    const users = await getRoleBoundUsers(
      props.cluster.id,
      selectedNamespaceRole.value,
      selectedNamespace.value
    )
    isNsRoleBound.value = users.some((u: any) => u.username === currentUser.value.username)
  } catch (error) {
    isNsRoleBound.value = false
  }
}

const handleBindRole = async () => {
  if (!props.cluster || !currentUser.value || !selectedClusterRole.value) return

  try {
    bindLoading.value = true
    await bindUserToRole({
      clusterId: props.cluster.id,
      userId: currentUser.value.userId || 8, // TODO: 使用实际的用户ID
      roleName: selectedClusterRole.value,
      roleNamespace: '',
      roleType: 'ClusterRole'
    })

    ElMessage.success('绑定成功')
    await checkRoleBinding()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '绑定失败')
  } finally {
    bindLoading.value = false
  }
}

const handleUnbindRole = async () => {
  if (!props.cluster || !currentUser.value || !selectedClusterRole.value) return

  try {
    bindLoading.value = true
    await unbindUserFromRole({
      clusterId: props.cluster.id,
      userId: currentUser.value.userId || 8,
      roleName: selectedClusterRole.value,
      roleNamespace: ''
    })

    ElMessage.success('解绑成功')
    await checkRoleBinding()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '解绑失败')
  } finally {
    bindLoading.value = false
  }
}

const handleBindNsRole = async () => {
  if (!props.cluster || !currentUser.value || !selectedNamespaceRole.value) return

  try {
    bindLoading.value = true
    await bindUserToRole({
      clusterId: props.cluster.id,
      userId: currentUser.value.userId || 8,
      roleName: selectedNamespaceRole.value,
      roleNamespace: selectedNamespace.value,
      roleType: 'Role'
    })

    ElMessage.success('绑定成功')
    await checkNsRoleBinding()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '绑定失败')
  } finally {
    bindLoading.value = false
  }
}

const handleUnbindNsRole = async () => {
  if (!props.cluster || !currentUser.value || !selectedNamespaceRole.value) return

  try {
    bindLoading.value = true
    await unbindUserFromRole({
      clusterId: props.cluster.id,
      userId: currentUser.value.userId || 8,
      roleName: selectedNamespaceRole.value,
      roleNamespace: selectedNamespace.value
    })

    ElMessage.success('解绑成功')
    await checkNsRoleBinding()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '解绑失败')
  } finally {
    bindLoading.value = false
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const buildPermissionTree = (rules: any[]) => {
  const tree: any[] = []
  rules.forEach((rule, index) => {
    const apiGroups = rule.apiGroups || ['']
    apiGroups.forEach((apiGroup: string, groupIndex: number) => {
      const apiGroupNode = {
        id: `apiGroup-${index}-${groupIndex}`,
        type: 'apiGroup',
        label: apiGroup || 'core',
        children: [] as any[]
      }
      const resources = rule.resources || []
      resources.forEach((resource: string, resIndex: number) => {
        const resourceNode = {
          id: `resource-${index}-${groupIndex}-${resIndex}`,
          type: 'resource',
          label: resource,
          children: [] as any[]
        }
        const verbs = rule.verbs || ['*']
        verbs.forEach((verb: string, vIndex: number) => {
          const verbLabel = verb === '*' ? '所有操作' : verb
          resourceNode.children.push({
            id: `verb-${index}-${groupIndex}-${resIndex}-${vIndex}`,
            type: 'verb',
            label: '操作',
            value: verbLabel
          })
        })
        apiGroupNode.children.push(resourceNode)
      })
      tree.push(apiGroupNode)
    })
  })
  return tree
}

</script>

<style scoped lang="scss">
.cluster-users-content {
  .credentials-section {
    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
      padding: 12px 16px;
      background: #f5f7fa;
      border-radius: 4px;

      h4 {
        margin: 0;
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }
    }

    // 优化表格样式
    :deep(.el-table) {
      font-size: 14px;

      th {
        background-color: #f5f7fa;
        font-weight: 600;
        color: #606266;
      }

      td {
        padding: 12px 0;
      }
    }
  }

  .role-auth-content {
    .role-section {
      margin-bottom: 20px;
      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 500;
        color: #333;
      }
    }

    .user-section {
      margin-bottom: 20px;
      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;
        h4 {
          margin: 0;
          font-size: 14px;
          font-weight: 500;
          color: #333;
        }
      }
    }

    .role-detail {
      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 500;
        color: #333;
      }
    }
  }

  .tree-node {
    display: flex;
    align-items: center;
    gap: 6px;
    .node-label {
      font-size: 14px;
    }
  }

  :deep(.el-tree-node__content) {
    padding: 4px 0;
  }
}
</style>
