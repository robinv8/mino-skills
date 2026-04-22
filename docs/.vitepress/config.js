import { defineConfig } from 'vitepress'

export default defineConfig({
    title: 'Mino Skills',
    description: 'Agent Skills compatible skill pack for task-driven development',
    base: '/mino-skills/',
    cleanUrls: true,

    head: [
        ['link', { rel: 'icon', type: 'image/png', href: '/logo.png' }]
    ],

    locales: {
        root: {
            label: 'English',
            lang: 'en',
            themeConfig: {
                nav: [
                    { text: 'Guide', link: '/guide/installation' },
                    { text: 'Skills', link: '/skills/task' },
                    { text: 'Reference', link: '/reference/iron-tree-protocol' },
                    { text: 'Advanced', link: '/advanced/loop-mode' },
                    { text: 'Changelog', link: '/migration/changelog' }
                ],

                sidebar: {
                    '/guide/': [
                        {
                            text: 'Getting Started',
                            items: [
                                { text: 'Introduction', link: '/' },
                                { text: 'Installation', link: '/guide/installation' },
                                { text: 'Quick Start', link: '/guide/quickstart' }
                            ]
                        }
                    ],
                    '/skills/': [
                        {
                            text: 'User Guide',
                            items: [
                                { text: 'mino-task', link: '/skills/task' },
                                { text: 'mino-run', link: '/skills/run' },
                                { text: 'mino-verify', link: '/skills/verify' },
                                { text: 'mino-checkup', link: '/skills/checkup' }
                            ]
                        }
                    ],
                    '/reference/': [
                        {
                            text: 'Protocol Reference',
                            items: [
                                { text: 'Iron Tree Protocol', link: '/reference/iron-tree-protocol' },
                                { text: 'Workflow State Contract', link: '/reference/workflow-state-contract' },
                                { text: 'Brief Contract', link: '/reference/brief-contract' }
                            ]
                        }
                    ],
                    '/advanced/': [
                        {
                            text: 'Advanced Topics',
                            items: [
                                { text: 'Loop Mode', link: '/advanced/loop-mode' },
                                { text: 'Composite Tasks', link: '/advanced/composite-tasks' },
                                { text: 'Manual Acceptance', link: '/advanced/manual-acceptance' },
                                { text: 'Adopt Existing Issues', link: '/advanced/adopt-existing-issues' },
                                { text: 'Silent Events (v1.10)', link: '/advanced/silent-events' },
                                { text: 'Brief Quality (v1.11+)', link: '/advanced/brief-quality' },
                                { text: 'Verification Report Artifact', link: '/advanced/verification-report' }
                            ]
                        }
                    ],
                    '/migration/': [
                        {
                            text: 'Migration',
                            items: [
                                { text: 'Changelog', link: '/migration/changelog' }
                            ]
                        }
                    ],
                    '/validation/': [
                        {
                            text: 'Validation',
                            items: [
                                { text: 'Phase 2 Regression Report', link: '/validation/phase2-regression' }
                            ]
                        }
                    ]
                }
            }
        },
        zh: {
            label: '简体中文',
            lang: 'zh-CN',
            link: '/zh/',
            themeConfig: {
                nav: [
                    { text: '指南', link: '/zh/guide/installation' },
                    { text: '技能', link: '/zh/skills/task' },
                    { text: '参考', link: '/zh/reference/iron-tree-protocol' },
                    { text: '进阶', link: '/zh/advanced/loop-mode' },
                    { text: '更新日志', link: '/zh/migration/changelog' }
                ],

                sidebar: {
                    '/zh/guide/': [
                        {
                            text: '快速开始',
                            items: [
                                { text: '介绍', link: '/zh/' },
                                { text: '安装', link: '/zh/guide/installation' },
                                { text: '快速上手', link: '/zh/guide/quickstart' }
                            ]
                        }
                    ],
                    '/zh/skills/': [
                        {
                            text: '用户指南',
                            items: [
                                { text: 'mino-task', link: '/zh/skills/task' },
                                { text: 'mino-run', link: '/zh/skills/run' },
                                { text: 'mino-verify', link: '/zh/skills/verify' },
                                { text: 'mino-checkup', link: '/zh/skills/checkup' }
                            ]
                        }
                    ],
                    '/zh/reference/': [
                        {
                            text: '协议参考',
                            items: [
                                { text: 'Iron Tree Protocol', link: '/zh/reference/iron-tree-protocol' },
                                { text: 'Workflow State Contract', link: '/zh/reference/workflow-state-contract' },
                                { text: 'Brief Contract', link: '/zh/reference/brief-contract' }
                            ]
                        }
                    ],
                    '/zh/advanced/': [
                        {
                            text: '进阶主题',
                            items: [
                                { text: 'Loop Mode', link: '/zh/advanced/loop-mode' },
                                { text: 'Composite Tasks', link: '/zh/advanced/composite-tasks' },
                                { text: 'Manual Acceptance', link: '/zh/advanced/manual-acceptance' },
                                { text: 'Adopt Existing Issues', link: '/zh/advanced/adopt-existing-issues' },
                                { text: 'Silent Events (v1.10)', link: '/zh/advanced/silent-events' },
                                { text: 'Brief Quality (v1.11+)', link: '/zh/advanced/brief-quality' },
                                { text: 'Verification Report Artifact', link: '/zh/advanced/verification-report' }
                            ]
                        }
                    ],
                    '/zh/migration/': [
                        {
                            text: '迁移',
                            items: [
                                { text: '更新日志', link: '/zh/migration/changelog' }
                            ]
                        }
                    ],
                    '/zh/validation/': [
                        {
                            text: '验证',
                            items: [
                                { text: 'Phase 2 回归测试报告', link: '/zh/validation/phase2-regression' }
                            ]
                        }
                    ]
                }
            }
        }
    },

    themeConfig: {
        logo: '/logo.png',

        search: {
            provider: 'local'
        },

        socialLinks: [
            { icon: 'github', link: 'https://github.com/robinv8/mino-skills' }
        ],

        editLink: {
            pattern: 'https://github.com/robinv8/mino-skills/edit/main/docs/:path'
        },

        footer: {
            message: 'Released under the MIT License.',
            copyright: 'Copyright © present'
        }
    }
})
