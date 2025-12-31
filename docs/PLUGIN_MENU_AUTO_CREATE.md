# 插件菜单自动创建机制

## 工作原理

当插件安装后，系统会自动在前端创建对应的菜单和路由。

### 完整流程

```
1. 应用启动
   ↓
2. main.ts 导入插件
   ↓
3. 插件自动注册到 pluginManager
   ↓
4. installPlugins() 自动安装所有插件
   ↓
5. 插件的 install() 方法被调用
   ↓
6. 插件路由注册到 Vue Router
   ↓
7. Layout 组件加载
   ↓
8. buildPluginMenus() 从 pluginManager 获取菜单
   ↓
9. buildMenuTree() 构建菜单树（根据 parentPath 和 sort）
   ↓
10. 菜单自动显示在侧边栏
```

## 关键代码

### 1. 插件定义菜单（kubernetes/index.ts）

```typescript
class KubernetesPlugin implements Plugin {
  // ... 其他代码 ...

  getMenus(): PluginMenuConfig[] {
    return [
      {
        name: 'Kubernetes管理',
        path: '/kubernetes',
        icon: 'Platform',
        sort: 100,           // 排序值，越小越靠前
        hidden: false,
        parentPath: '',      // 空表示顶级菜单
      },
      {
        name: '集群管理',
        path: '/kubernetes/clusters',
        icon: 'OfficeBuilding',
        sort: 1,
        hidden: false,
        parentPath: '/kubernetes',  // 父菜单路径
      },
      // ... 其他子菜单
    ]
  }
}
```

### 2. 插件自动安装（main.ts）

```typescript
// 导入插件（插件会自动注册到 pluginManager）
import '@/plugins/kubernetes'

// 自动安装所有已注册的插件
async function installPlugins() {
  const plugins = pluginManager.getAll()
  for (const plugin of plugins) {
    await pluginManager.install(plugin.name)
  }
}

// 安装插件并注册路由
installPlugins().then(() => {
  registerPluginRoutes()
  app.use(router)
  app.use(ElementPlus)
  app.mount('#app')
})
```

### 3. 菜单自动构建（Layout.vue）

```typescript
// 从插件管理器构建菜单
const buildPluginMenus = () => {
  const pluginMenus: any[] = []
  const plugins = pluginManager.getInstalled()  // 获取已安装的插件

  plugins.forEach(plugin => {
    if (plugin.getMenus) {
      const menus = plugin.getMenus()
      menus.forEach(menu => {
        pluginMenus.push({
          ID: menu.path,
          name: menu.name,
          path: menu.path,
          icon: menu.icon,
          sort: menu.sort,
          hidden: menu.hidden,
          parentPath: menu.parentPath,
          children: []
        })
      })
    }
  })

  return pluginMenus
}

// 构建菜单树
const buildMenuTree = (menus: any[]) => {
  const menuMap = new Map()
  menus.forEach(menu => {
    menuMap.set(menu.path, { ...menu, children: [] })
  })

  const tree: any[] = []

  // 根据 parentPath 构建树结构
  menus.forEach(menu => {
    const menuItem = menuMap.get(menu.path)
    if (menu.parentPath && menuMap.has(menu.parentPath)) {
      // 有父菜单，添加到父菜单的 children
      const parent = menuMap.get(menu.parentPath)
      parent.children.push(menuItem)
    } else {
      // 没有父菜单，添加到根节点
      tree.push(menuItem)
    }
  })

  // 按 sort 排序
  const sortMenus = (menus: any[]) => {
    menus.sort((a, b) => (a.sort || 0) - (b.sort || 0))
    menus.forEach(menu => {
      if (menu.children && menu.children.length > 0) {
        sortMenus(menu.children)
      }
    })
  }

  sortMenus(tree)
  return tree
}

// 加载菜单
const loadMenu = async () => {
  // 1. 获取系统菜单
  let systemMenus: any[] = []
  try {
    systemMenus = await getUserMenu() || []
  } catch (error) {
    console.log('获取系统菜单失败，仅显示插件菜单:', error)
  }

  // 2. 获取插件菜单
  const pluginMenus = buildPluginMenus()

  // 3. 合并菜单
  const allMenus = [...systemMenus, ...pluginMenus]

  // 4. 构建菜单树
  menuList.value = buildMenuTree(allMenus)
}
```

## 菜单配置说明

### PluginMenuConfig 接口

```typescript
export interface PluginMenuConfig {
  name: string         // 菜单名称
  path: string         // 菜单路径（与路由路径一致）
  icon: string         // 图标名称（Element Plus 图标）
  sort: number         // 排序值（越小越靠前）
  hidden: boolean      // 是否隐藏
  parentPath: string   // 父菜单路径（空字符串表示顶级菜单）
  permission?: string  // 权限标识（可选）
}
```

### 菜单层级示例

```typescript
// Kubernetes 管理插件的菜单结构
[
  {
    name: 'Kubernetes管理',      // 顶级菜单
    path: '/kubernetes',
    parentPath: '',              // 空表示顶级
    sort: 100,
    children: [
      {
        name: '集群管理',        // 子菜单
        path: '/kubernetes/clusters',
        parentPath: '/kubernetes',  // 父菜单路径
        sort: 1
      },
      {
        name: '节点管理',
        path: '/kubernetes/nodes',
        parentPath: '/kubernetes',
        sort: 2
      },
      // ... 更多子菜单
    ]
  }
]
```

## 菜单显示效果

安装 Kubernetes 插件后，侧边栏会自动显示：

```
📊 首页
👤 系统管理
   ├─ 用户管理
   ├─ 角色管理
   ├─ 部门管理
   └─ 菜单管理
🖥️ Kubernetes管理        ← 插件自动添加
   ├─ 集群管理            ← 子菜单自动生成
   ├─ 节点管理
   ├─ 工作负载
   ├─ 命名空间
   ├─ 网络管理
   ├─ 配置管理
   ├─ 存储管理
   ├─ 访问控制
   ├─ 终端审计
   └─ 应用诊断
```

## 动态控制

### 启用插件 - 菜单自动出现

```typescript
import { pluginManager } from '@/plugins/manager'

// 安装插件（菜单会自动出现）
await pluginManager.install('kubernetes')

// 刷新页面后菜单显示
location.reload()
```

### 卸载插件 - 菜单自动消失

```typescript
// 卸载插件（菜单会自动消失）
await pluginManager.uninstall('kubernetes')

// 刷新页面后菜单隐藏
location.reload()
```

## 注意事项

1. **插件必须先安装** - 只有已安装的插件（通过 `pluginManager.install()`）菜单才会显示
2. **需要刷新页面** - 启用/禁用插件后需要刷新页面才能看到菜单变化
3. **图标必须注册** - 使用的图标必须在 Layout.vue 的 iconMap 中注册
4. **路径必须唯一** - 菜单的 path 必须是唯一的，不能重复
5. **父菜单必须存在** - 如果指定了 parentPath，父菜单必须存在

## 调试技巧

### 1. 查看已安装的插件

```typescript
const plugins = pluginManager.getInstalled()
console.log('已安装的插件:', plugins)
```

### 2. 查看插件菜单配置

```typescript
const plugin = pluginManager.get('kubernetes')
if (plugin && plugin.getMenus) {
  console.log('插件菜单:', plugin.getMenus())
}
```

### 3. 查看最终菜单树

在浏览器控制台查看：
```
插件菜单: [...]
最终菜单树: [...]
```

## 扩展开发

### 添加新的菜单项

如果你想给 Kubernetes 插件添加新的菜单项，只需修改 `web/src/plugins/kubernetes/index.ts`:

```typescript
getMenus(): PluginMenuConfig[] {
  return [
    // ... 现有菜单 ...
    {
      name: '新功能',
      path: '/kubernetes/newfeature',
      icon: 'Star',
      sort: 11,  // 放在最后
      hidden: false,
      parentPath: '/kubernetes',
    },
  ]
}
```

### 添加三级菜单

```typescript
{
  name: '高级功能',
  path: '/kubernetes/advanced',
  icon: 'Setting',
  sort: 11,
  hidden: false,
  parentPath: '/kubernetes',  // 二级菜单
},
{
  name: '高级配置',
  path: '/kubernetes/advanced/config',
  icon: 'Tools',
  sort: 1,
  hidden: false,
  parentPath: '/kubernetes/advanced',  // 三级菜单
}
```

## 总结

插件安装后，菜单会**自动**：
1. ✅ 从插件的 `getMenus()` 方法获取配置
2. ✅ 根据 `parentPath` 构建树形结构
3. ✅ 根据 `sort` 值排序
4. ✅ 在侧边栏渲染显示
5. ✅ 与系统菜单合并显示

**无需手动配置，一切自动完成！** 🎉
