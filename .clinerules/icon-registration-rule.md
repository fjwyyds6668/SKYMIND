# 图标注册规则 (Icon Registration Rule)

## 规则描述
在项目中使用 lucide-vue-next 图标时，必须遵循以下注册规则以防止组件名称冲突。

## 规则内容

### 1. 全局注册规则
在 `frontend/src/main.js` 中，所有 lucide-vue-next 图标必须以 "Icon" 前缀进行全局注册：

```javascript
// 全局注册 lucide-vue-next 图标
import * as Icons from 'lucide-vue-next';

// 全局注册所有图标组件，添加 Icon 前缀防止冲突
Object.entries(Icons).forEach(([name, component]) => {
  const iconName = `Icon${name}`;
  app.component(iconName, component);
});
```

### 2. 使用规则
在任何 Vue 组件中使用图标时，必须使用带 "Icon" 前缀的组件名：

```vue
<template>
  <!-- 正确：使用带前缀的图标名 -->
  <IconChevronDown />
  <IconSend />
  <IconPlus />
  <IconX />
  
  <!-- 错误：不要使用原始图标名 -->
  <!-- <ChevronDown /> -->
  <!-- <Send /> -->
</template>
```

### 3. 常用图标映射
| 原始图标名 | 项目中使用名 |
|-----------|-------------|
| ChevronDown | IconChevronDown |
| ChevronUp | IconChevronUp |
| Send | IconSend |
| Plus | IconPlus |
| X | IconX |
| Menu | IconMenu |
| Search | IconSearch |
| User | IconUser |
| Settings | IconSettings |
| Bell | IconBell |
| Home | IconHome |
| MessageCircle | IconMessageCircle |
| Phone | IconPhone |
| Mail | IconMail |
| Calendar | IconCalendar |
| Clock | IconClock |
| Star | IconStar |
| Heart | IconHeart |
| Share2 | IconShare2 |
| Download | IconDownload |
| Upload | IconUpload |
| Edit | IconEdit |
| Trash2 | IconTrash2 |
| Eye | IconEye |
| EyeOff | IconEyeOff |
| Check | IconCheck |
| AlertCircle | IconAlertCircle |
| Info | IconInfo |
| HelpCircle | IconHelpCircle |
| ArrowLeft | IconArrowLeft |
| ArrowRight | IconArrowRight |
| ArrowUp | IconArrowUp |
| ArrowDown | IconArrowDown |

## 应用场景
- 当在 Vue 组件中添加新的 lucide-vue-next 图标时
- 当重构现有代码，发现未使用 "Icon" 前缀的图标时
- 当进行代码审查，检查图标命名规范时

## 注意事项
1. 所有 lucide-vue-next 图标都必须使用 "Icon" 前缀
2. 前缀后的第一个字母必须大写（PascalCase）
3. 不要在组件中直接导入单个图标，使用全局注册的组件
4. 如果发现项目中存在未使用前缀的图标，需要立即修正

## 示例修正
**修正前：**
```vue
<template>
  <ChevronDown />
  <Send />
</template>
```

**修正后：**
```vue
<template>
  <IconChevronDown />
  <IconSend />
</template>
