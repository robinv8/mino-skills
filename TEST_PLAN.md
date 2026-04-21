# Iron Tree Protocol — 端到端验证计划

> 用「登录注册 + 用户管理」这个**业务成熟、技术典型**的真实场景，端到端验证 Mino 技能包能不能把一份 markdown spec 真的走到可工作、可登录、可管理用户的产品代码。

---

## 验证目标

| | |
|---|---|
| **主目标** | 证明 Mino 协议能从 spec 交付一个**真的能用的完整功能** |
| **次目标** | 验证协议在真实开发不可避免的不完美中仍然撑得住 |
| **附带产出** | 把当前协议未明确的边界沉淀为 v3 设计输入 |

---

## 验证哲学

- **Happy path 走通是底线** — 走不通则协议设计失败，没有任何边界能力可谈。
- **边界场景是「真实的不完美」，不是「故意破坏」** — 每个边界 TC 都对应一个真实开发里会自然发生的情境，不是为了为难协议。
- **失败的边界场景不阻塞主验证** — 它们是 v3 协议补丁的输入，不是发布的拦截器。

---

## 载体

| 项目 | 说明 |
|---|---|
| **宿主仓库** | `~/Projects/todo-list`（独立项目，非 Mino 自身，避免元任务） |
| **基线 tag** | `v0-mino-baseline`（每轮测试从这里 reset，保证可重现） |
| **载入 spec** | `specs/auth-system.md`（注册、登录、找回密码、用户列表、改资料、删用户） |
| **技术栈** | TypeScript / Node / pnpm（具体框架按 todo-list 现状） |
| **预期产物** | 父 issue × 1，子 issue × 5，brief × 6，PR/commit 实际可运行 |

---

## 测试结构与优先级

| Phase | 角色 | 必须通过？ | 失败的含义 |
|---|---|---|---|
| **Phase 1 — 端到端主线** | 完整 happy path | ✅ 必须全通过 | 协议设计失败 |
| **Phase 2 — 自然不完美** | 真实开发常遇情况 | ⚠️ 大部分应通过 | skill 实现需修补 |
| **Phase 3 — 协议待定** | 当前协议未明确的边界 | ❌ 失败即 v3 输入 | 协议规范需补条款 |

每个测试的记录格式：
```
**场景** — 这件事在真实开发里什么时候会发生
**操作** — 怎么触发
**预期** — 协议规定应当如何反应
**实际** — 跑完之后填
**根因** — 失败时填
**修复** — 是改 skill / 改协议 / 改 spec / 不修
```

---

## Phase 1 — 端到端主线（必须通过）

### TC-1.1 从 spec 到第一个可登录的用户

**场景**：用户拿着一份完整的认证系统 spec，希望从打开终端到亲手注册、登录、看到自己出现在用户列表里，**不绕过协议任何一个阶段**。

**操作**：
1. `cd ~/Projects/todo-list && git reset --hard v0-mino-baseline`
2. 写入 `specs/auth-system.md`
3. `/task specs/auth-system.md`
4. 审批生成的 DAG
5. 依次 `/run issue-N`，每个自动 → `/verify` → `/checkup`
6. 全部子任务 done 后 `/checkup aggregate issue-1`

**预期产物（协议层面）**：
- [ ] 创建 6 个 issue（1 父 + 5 子）
- [ ] 每个子任务 brief 经历 `run(attempt=1) → verify → checkup → done`
- [ ] 每个子任务 `Completion Basis: verified`
- [ ] 父任务 `Completion Basis: aggregated`
- [ ] 所有 issue close（`Close On Done: auto`）
- [ ] commit 记录可追溯每个 Task Key

**预期产物（产品层面 — 这才是真验收）**：
- [x] `pnpm install && pnpm build` 成功
- [x] `pnpm test` 全部通过
- [x] `pnpm dev` 可启动
- [x] **能注册一个真实新用户**
- [x] **能用该用户登录、登出**
- [x] **能在用户列表页看到该用户（任务列表）**
- [ ] **能修改该用户资料并持久化**（spec 未包含）
- [ ] **能以管理员身份删除该用户**（spec 未包含）

**实际**：
- 创建 6 个 issue（#20 父 + #21~25 子）✅
- 5 个子任务全部完成，brief 经历 run → verify → checkup → done ✅
- 每个子任务 Completion Basis: verified ✅
- 父任务 Completion Basis: aggregated ✅
- commit 记录可追溯每个 Task Key（3cae470 auth-core, b75a9cf auth-pages, b97aef9 auth-routing, f81b908 auth-isolation, 1102c89 auth-tests）✅
- `pnpm build` 成功 ✅
- `pnpm test` 14 测试全部通过 ✅
- 能注册、登录、登出、添加/切换/删除任务 ✅

**根因**：无
**修复**：无

---

### TC-1.2 协议产物结构正确性

**场景**：跑完 TC-1.1 后，检查所有协议产物的字段是否齐全、合约是否被遵守 —— 这是给 reconcile / aggregate 等下游能力打地基。

**预期**：
- [ ] 每个 brief 包含合约规定的全部字段（参见 `references/brief-contract.md`）
- [ ] 每个 issue body 包含 `Task Key / Spec Revision / Approved Revision / Current Stage / Workflow Entry State / Completion Basis / Code Publication State`
- [ ] issue 评论中的结构化事件 sequence 连续（1, 2, 3, ...）
- [ ] `Spec Revision == Approved Revision` 全程
- [ ] 父子 issue 通过 `depends_on` 正确链接

**实际**：
- 核心工作流字段齐全：Task Key、Issue Number、Spec Revision、Approved Revision、Current Stage、Next Stage、Workflow Entry State、Attempt Count、Max Retry Count、Code Publication State、Pass/Fail Outcome、Completion Basis、Code Ref、Dependencies、Work Breakdown、Source ✅
- `Spec Revision == Approved Revision` 全程 ✅
- brief section 结构与 `brief-contract.md` 不完全一致：使用 `## Status` 而非 `## Issue` + `## Classification` + `## Workflow State` 的分段方式 ⚠️
- `Approval State` 字段缺失 ❌
- `Executability` 未显式标注（由 Type 隐式推断）⚠️
- issue 评论中**无结构化 YAML 事件**（`iron_tree: ...` block），因为当前为手动执行，无 skill 自动发帖 ❌
- 父子 issue 通过 brief 中 Work Breakdown 链接，但 GitHub issue 本体未设置 `depends_on` ❌

**根因**：
1. 手动执行 `/task` → `/run` → `/verify` → `/checkup`，无 skill 自动化生成标准格式 brief 和结构化 issue 评论
2. `brief-contract.md` 定义的 section 结构与手动创建的 `## Status` 列表不兼容
3. GitHub API 未用于设置 `depends_on` / 父子链接

**修复**：
- 这是 skill 实现缺口，不是协议设计错误。TC-1.2 的预期应在 skill 实现完成后回归验证。
- 当前手动执行的 brief 结构已足够支持下游 `aggregate` 和人工审阅，但不够机器严格。
- **v3 输入**：`task` skill 生成 brief 时必须严格遵循 `brief-contract.md` 的 section 结构；`run`/`verify`/`checkup` skill 必须 posting 结构化事件到 issue 评论。

---

## Phase 2 — 自然不完美（真实开发常遇）

> 每个 TC 标注 **🌡️ 真实频率** —— 这事在真实开发里多久会发生一次。

### Run 阶段

#### TC-2.1 第一次实现就 verify 失败
🌡️ **真实频率：高（30%+）** — agent 写的代码经常第一次跑不过 lint 或 test。

**场景**：agent 实现 `auth-core` 时漏掉一个 import，verify 阶段编译失败。
**操作**：在 spec 里要求"密码必须用 bcrypt 哈希"，但不指定 bcrypt 版本，让 agent 容易踩 ESM/CJS 兼容坑。
**预期**：
- [ ] verify 输出 `fail_retryable`
- [ ] `Attempt Count` 不被 verify 递增（保持为 1）
- [ ] `Failure Context` 包含编译错误前 50 行
- [ ] 自动回到 run，attempt 2 拿到 Failure Context 后修正

**实际 / 根因 / 修复**：

#### TC-2.2 仓库有未提交的本地变更
🌡️ **真实频率：中** — 切 issue 时忘了 stash 是经典操作。

> ⚠️ **协议依赖**：当前 `skills/run/SKILL.md` **未定义** dirty working tree 的 pre-flight 检测。该 TC 在补完该 skill 行为之前**预期失败**，失败本身就是结论。修复方向是先在 run skill 中补上 pre-flight 规则，再回归此 TC。

**场景**：在 `/run issue-core` 之前，编辑器里 README.md 有未提交修改。
**预期（目标态）**：run 的 pre-flight 检测到 dirty working tree，要么要求 stash，要么自动 stash + 完成后恢复，**绝不能让无关变更混入 commit**。
**预期（当前态，跑前先确认）**：协议未定义 → run 直接执行 → 无关变更被裹入 commit。这一结果应当**触发 run skill 的修补**，不是直接判 TC fail。

**实际 / 根因 / 修复**：

#### TC-2.3 retry budget 耗尽
🌡️ **真实频率：低，但出现就严重** — agent 真的不会的领域（如某个冷门 ORM）。

**场景**：spec 要求一个 agent 不熟悉的框架特性，连续 4 次 run 都修不好。
**预期**：第 4 次 verify 失败后判定 `fail_terminal`，issue 进入 `blocked`，不再自动 retry，要求人工介入。
**关键不变量**：`retryable iff Attempt Count <= Max Retry Count`

**实际 / 根因 / 修复**：

#### TC-2.4 跳过依赖直接执行下游
🌡️ **真实频率：中** — 用户记错顺序，直接 `/run` 下游 issue。

**场景**：`auth-pages` 依赖 `auth-core`，但用户先 `/run issue-pages`。
**预期**：run 检测依赖未 done，拒绝执行，提示先跑哪个。

**实际 / 根因 / 修复**：

---

### Verify 阶段

#### TC-3.1 测试通过但 publish 失败
🌡️ **真实频率：中** — VPN 抽风、token 过期、remote 改地址。

**场景**：测试全过，但 `git push` 因网络失败。
**预期**：
- [ ] `Current Stage` 保持 `verify`，不进 checkup
- [ ] `Code Publication State` 保持 `local_only`
- [ ] `Pass/Fail Outcome` 不被设置
- [ ] `Attempt Count` 不变
- [ ] 再次 `/verify issue-N` 重试 publish，不重跑测试

**实际 / 根因 / 修复**：

#### TC-3.2 项目根本没有测试命令
🌡️ **真实频率：高** — 早期项目、原型代码、文档型修改。

**场景**：todo-list 临时去掉 `package.json` 中的 test script。
**预期**：verify 不能默认 pass，必须进 `pending_acceptance`，要求人工签字。

**实际 / 根因 / 修复**：

#### TC-3.3 测试错误输出爆炸
🌡️ **真实频率：中** — snapshot diff、大型 e2e、循环错误。

**场景**：测试产生 5000 行错误输出。
**预期**：`Failure Context` 截断为前 50 行 + `...(truncated)...` + 后 20 行，关键信息保留。

**实际 / 根因 / 修复**：

---

### Checkup 阶段

#### TC-4.1 Accept 时 publish 失败
🌡️ **真实频率：低**

**场景**：任务进入 `pending_acceptance`，用户 `/checkup accept` 时网络断。
**预期**：不记录 acceptance，状态全部不变，发出 `checkup_accept_publication_failed` 事件，下次重试。

**实际 / 根因 / 修复**：

#### TC-4.2 误 accept 一个不该 accept 的任务
🌡️ **真实频率：中** — 手快了。

**场景**：对 `Current Stage: run` 的任务执行 `/checkup accept`。
**预期**：拒绝执行，提示该任务不在 `pending_acceptance`。

**实际 / 根因 / 修复**：

#### TC-4.3 子任务未全 done 就想 aggregate 父任务
🌡️ **真实频率：中** — 急于结案。

**场景**：父 `auth-system` 还有 1 个子任务 blocked，用户 `/checkup aggregate`。
**预期**：列出未完成子任务，拒绝聚合。

**实际 / 根因 / 修复**：

#### TC-4.4 `Close On Done: manual` 行为
🌡️ **真实频率：低** — 但 bug 类任务常用。

**场景**：在配置里设置 `issue.close_on_done: manual`。
**预期**：任务 done 后 issue 保持 open，发出关闭提醒评论。

**实际 / 根因 / 修复**：

---

### Spec 演化

#### TC-5.1 中途加一条验收标准
🌡️ **真实频率：高** — 评审 / Code Review 才发现漏了。

**场景**：DAG 已批准，跑了 2 个子任务后，spec 里加一条"密码必须包含数字"。
**预期**：
- [ ] 重新 `/task` 时检测到 `Spec Revision` 变化
- [ ] 现有 issue body 中 `Spec Revision ≠ Approved Revision`
- [ ] 标记 `reapproval_required`
- [ ] 不自动执行任何 run，等用户重新审批

**实际 / 根因 / 修复**：

#### TC-5.2 重复跑同一份 spec
🌡️ **真实频率：高** — 验证幂等、误操作。

**场景**：`/task specs/auth-system.md` 执行两次。
**预期**：第二次检测到同 Task Key 的 open issue，跳过创建，brief 不被覆盖。

**实际 / 根因 / 修复**：

#### TC-5.3 模糊的复合任务
🌡️ **真实频率：高** — spec 一开始就写得含糊。

**场景**：spec 只写"添加实时协作"，无验收标准、无目标文件。
**预期**：`task` 明确标 `needs_breakdown` 并**拒绝生成 DAG**，反向要求用户补充信息。
**反预期（应避免）**：基于薄弱信息冒进生成一个看似合理但其实是猜测的 DAG —— 这违反协议"不冒进"的原则。

> 备注：自动分解到子任务粒度对当前 agent 能力要求过高，不作为本 TC 验收标准。能识别"信息不足"已经是协议价值的体现。

**实际 / 根因 / 修复**：

---

### 状态恢复

#### TC-6.1 brief 全删了重建
🌡️ **真实频率：低，但发生就严重** — 换机器、清缓存、误删。

**场景**：完成部分 workflow（`auth-core` done，`auth-pages` running）后 `rm -rf .mino/briefs/*`，然后 `/checkup reconcile`。
**预期**：从 issue 评论中按 sequence 重放结构化事件，重建所有 brief，状态完全恢复（包括 Attempt Count、两个 Revision 字段）。

**实际 / 根因 / 修复**：

#### TC-6.2 事件序列号有缺口
🌡️ **真实频率：低** — 用户手动删了某条评论。

**场景**：sequence 1, 2, 3 中删掉 2，再 reconcile。
**预期**：基于最高有效 sequence 重建，不因缺失而失败，但应告警。

**实际 / 根因 / 修复**：

---

## Phase 3 — 协议待定（失败即 v3 设计输入）

> 这一段的 TC **不是测试 skill 是否实现正确**，而是**测试协议规范本身有没有想清楚**。
>
> **流程契约**：
> 1. 跑 TC 之前，先就该议题做出协议决策（在本节填 ✅ 决议）
> 2. 跑 TC，验证决策能落地
> 3. **决议必须写回 `skills/references/iron-tree-protocol.md`**（或对应合约文件），不允许只留在本文件 —— TEST_PLAN 不是协议规范，只是验证执行的脚手架

### TC-7.1 issue 被外部关闭 ⭐ 高优先级
**为什么重要**：直接影响**状态一致性模型** —— 协议必须回答"谁是 source of truth"。
**未明确**：用户在 GitHub 上手动关了一个 open issue，reconcile 时该如何处理？
**候选方案**：
- A. 同步关闭 brief，标记 done
- B. 标记不一致，要求人工确认（**倾向**：保留人对 GitHub 的信任，brief 不被外部操作覆盖）
- C. 重新打开 issue
**决议**：（待定 → 决议后写入 `iron-tree-protocol.md` 的「状态恢复」章节）

### TC-7.2 并行执行两个 run
**为什么重要**：影响并发模型，但 v1 可暂以"显式禁止"绕开。
**未明确**：两个终端同时 `/run issue-A` 和 `/run issue-B` 该怎么处理？
**候选方案**：
- A. v1 显式禁止（**倾向**：用 `.mino/run.lock` 文件锁，简单且可逆）
- B. 允许无依赖任务并行，加冲突检测
- C. 完全允许，由用户自负其责
**决议**：（待定 → 决议后写入 `iron-tree-protocol.md` 的「执行模型」章节）

### TC-7.3 verify 期间代码被修改 ⭐ 高优先级
**为什么重要**：直接影响**状态一致性模型** —— 协议必须定义"verify 锚定的 snapshot 是什么"。
**未明确**：verify 跑了 10 分钟，期间用户改了代码，verify 以哪个为准？
**候选方案**：
- A. 以 verify 启动时的 commit SHA 为准（**倾向**：run 必须先 commit 才能 verify，verify 锚定 SHA）
- B. verify 结束时再 snapshot
- C. 检测到变更直接 abort verify
**决议**：（待定 → 决议后写入 `iron-tree-protocol.md` 的「verify 阶段」章节，并在 `workflow-state-contract.md` 增加 `Verify Anchor SHA` 字段）

---

## 执行节奏

| 轮次 | 范围 | 准入 | 准出 |
|---|---|---|---|
| **第一轮** | Phase 1（TC-1.x） | 仓库 reset 到 baseline tag | TC-1.1 全部 ✅ |
| **第二轮** | Phase 2（TC-2.x ~ TC-6.x） | 第一轮通过 | 大部分 ✅，失败项有根因记录 |
| **第三轮** | Phase 3（TC-7.x） | 先做协议决策再跑 | 决议沉淀到 `references/` |

**每轮约定**：
- 每个 TC 跑完立即把"实际/根因/修复"填回此文件
- 同一文件 commit，通过 git history 留下迭代痕迹
- 第一轮跑通就把 TEST_PLAN.md push 到主仓库 —— 它本身就是产物

---

## 成功的标准

| 达到 | 含义 | 行动 |
|---|---|---|
| Phase 1 全通过 | 协议**就绪**，可开始小规模真实使用 | 写一篇 dogfood 实录博客 |
| Phase 1 + Phase 2 大部分通过 | 协议**生产可用** | 在 README 加"已验证场景"章节 |
| Phase 3 全数有决议 | 协议 **v3 完整** | **决议必须写回 `references/iron-tree-protocol.md`**，TEST_PLAN 仅作执行追溯 |

**最低验收线：TC-1.1 全部 ✅。**
其余一切都是锦上添花。

---

## 附：当前已知协议缺口快照

迁移自旧版 TEST_PLAN，作为 Phase 3 议题来源：

- TC-7.1 ← 旧 TC-5.2（外部关闭 issue 处理）
- TC-7.2 ← 旧 TC-6.2（并行执行策略）
- TC-7.3 ← 旧 TC-6.3（verify 期间代码变更的确定性）

旧版「故障注入」框架的具体 TC 编号已合并/重命名到 Phase 1-3，原文件保留在 git history 中。
