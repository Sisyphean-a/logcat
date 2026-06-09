import { main } from "../wailsjs/go/models";

export function createMockState() {
  return new main.AppState({
    status: "running",
    adbStatus: "已连接",
    devices: [
      { id: "SM-A217F", model: "SM-A217F", status: "device" },
    ],
    selectedDevice: "SM-A217F",
    packageScope: "all",
    packages: [
      { name: "xxx.hostapp" },
      { name: "xxx.hostapp.dev" },
      { name: "com.android.systemui" },
    ],
    selectedPackage: "xxx.hostapp",
    processes: [
      { pid: 2001, name: "xxx.hostapp" },
      { pid: 2002, name: "xxx.hostapp:webview" },
    ],
    selectedProcess: "",
    boundPids: [2001, 2002],
    totalLogs: 16,
    visibleCount: 16,
    visibleStart: 0,
    filter: {
      draft: "",
      applied: "",
      error: "",
      activeFilterId: "",
      saved: [],
      history: [],
    },
    search: {
      query: "",
      matchIndexes: [],
      current: -1,
    },
    pause: {
      active: false,
      bufferedCount: 0,
      droppedCount: 0,
    },
    selectedIndex: -1,
    logs: [
      logRow(0, "16:42:18.479", "I", "chromium", "[H5] 进入申请页面", "views/apply/index.vue:12"),
      logRow(1, "16:42:19.120", "I", "chromium", "[H5] 请求 /api/apply/start", "api/request.ts:44"),
      logRow(2, "16:42:19.800", "W", "chromium", "[H5] token 即将过期，剩余 180s", "utils/auth.ts:88"),
      logRow(3, "16:42:20.411", "E", "chromium", "[H5] TypeError: Cannot read properties of undefined (reading 'productId')", "views/apply/index.vue:57"),
      logRow(4, "16:42:20.412", "E", "chromium", "at computed (index.vue:57) at ReactiveEffect.run (reactivity.js:185)", ""),
      logRow(5, "16:42:21.100", "I", "chromium", "[H5] 挂载 ProductApply 组件", "views/apply/index.vue:5"),
      logRow(6, "16:42:22.100", "I", "chromium", "[H5] 获取用户资料成功 { userId: 'U_88712', name: '李明' }", "api/user.ts:31"),
      logRow(7, "16:42:22.780", "W", "chromium", "[H5] 性能警告: LCP 2.3s, 建议 < 2.5s", "utils/perf.ts:19"),
      logRow(8, "16:42:23.100", "I", "chromium", "[H5] 用户点击提交按钮", "views/apply/index.vue:101"),
      logRow(9, "16:42:23.450", "E", "chromium", "[H5] 网络错误: 请求超时 POST /api/apply/submit (30s)", "api/request.ts:112"),
      logRow(10, "16:42:24.100", "I", "chromium", "[H5] 重试请求 /api/apply/submit (attempt 1/3)", "api/request.ts:88"),
      logRow(11, "16:42:25.200", "I", "chromium", "[H5] 提交成功，跳转到结果页 /apply/result?id=TXN_20240101", "views/apply/index.vue:134"),
      logRow(12, "16:42:26.100", "I", "chromium", "[H5] 结果页加载完成，耗时 320ms", "views/result/index.vue:8"),
      logRow(13, "16:42:27.020", "I", "chromium", "[H5] 打开帮助页 https://example.com/help/apply?from=h5", "router/index.ts:42"),
      logRow(14, "16:42:28.550", "I", "chromium", "at renderSubmit (views/apply/index.vue:144)", ""),
      logRow(15, "16:42:29.180", "I", "chromium", "[H5] DELETE /api/apply/draft/42 完成，耗时 188ms", "api/request.ts:151"),
    ],
  });
}

function logRow(
  index: number,
  timeText: string,
  level: string,
  tag: string,
  message: string,
  source: string,
) {
  return {
    index,
    timeText,
    level,
    tag,
    message,
    source,
    raw: `${timeText} ${level} ${tag} ${message}`,
    display: `${timeText} ${level} ${tag} ${message}`,
    isMatch: false,
    isCurrent: false,
    isSelected: false,
  };
}
