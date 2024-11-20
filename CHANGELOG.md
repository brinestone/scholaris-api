# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## 2.0.0 (2024-11-20)


### Features

* **auth:** account deletion endpoint ([6cfe576](https://github.com/brinestone/scholaris-api/commit/6cfe5761b3130e454a792f4ac16074c7b13d413e))
* **forms:** creation of form questions ([1fa155a](https://github.com/brinestone/scholaris-api/commit/1fa155a1c96cf078825fa0c6cbafa45b44a84d6d))
* **forms:** CRUD endpoints implemented ([4867866](https://github.com/brinestone/scholaris-api/commit/486786693d6e3755e1fba2fa6595060f773b85a9))
* **forms:** deletion of form groups ([ebad11c](https://github.com/brinestone/scholaris-api/commit/ebad11c0acbb292ecba563c40cacbf7b43786d9e))
* **forms:** fetch form info ([8de36d6](https://github.com/brinestone/scholaris-api/commit/8de36d616435e03edd572d1ffecdeb2e8538f76b))
* **forms:** form answer endpoints ([e0ce62d](https://github.com/brinestone/scholaris-api/commit/e0ce62d890365dc7e5399de46209857888fafd8f))
* **forms:** form question update support ([683372b](https://github.com/brinestone/scholaris-api/commit/683372b49a94d92b777409d77aa02559a71f8fb0))
* **forms:** form status toggling ([7f2854c](https://github.com/brinestone/scholaris-api/commit/7f2854c16033acc242920e7d24f73f4bc9a96225))
* **forms:** question creation ([554175f](https://github.com/brinestone/scholaris-api/commit/554175fffec08cb2bca5db0a9985b3d87bbac890))
* **institution:** enrollment creation ([75df58b](https://github.com/brinestone/scholaris-api/commit/75df58bbb66e185f56a5134aa876ddb25a75b6ae))
* **institutions:** academic year creation ([2783b18](https://github.com/brinestone/scholaris-api/commit/2783b18eb803d145a140bd7adc52f6c2f990c9cb))
* **institutions:** create new, lookup and get info ([6f8c5bf](https://github.com/brinestone/scholaris-api/commit/6f8c5bf91edbe7ce686902583d088882a5d90291))
* **institutions:** new enrollment form creation ([50686f0](https://github.com/brinestone/scholaris-api/commit/50686f06a14eac3ed87c688093c3b84301431de3))
* **institutions:** new-enrollments ([d8652b5](https://github.com/brinestone/scholaris-api/commit/d8652b51887b392a31c6e9d5b6ed016c5d1c62d7))
* new form support ([58027bb](https://github.com/brinestone/scholaris-api/commit/58027bb8a90c427a878bc181e28b70d71ea06e30))
* **settings:** add/update and read endpoints ([12e3e6c](https://github.com/brinestone/scholaris-api/commit/12e3e6c9a3351a8adfb7da5744c2e1bed3f282ea))
* **settings:** value update endpoint ([411a978](https://github.com/brinestone/scholaris-api/commit/411a978cb6be101a2cefb10dc55483cceccb0f9a))
* **users:** default user avatar ([8a77818](https://github.com/brinestone/scholaris-api/commit/8a77818aba9631df38f0eb8c1dc16f5e7a95061d))


### Bug Fixes

* **app:** configure cors for staging-development communication ([b9e2481](https://github.com/brinestone/scholaris-api/commit/b9e2481695a9751e7ce99e040c432317a319cc8d))
* **app:** cors error ([5349878](https://github.com/brinestone/scholaris-api/commit/53498788f6b06df4b8e2ae5250d94207a122da07))
* **app:** cors error ([4331c1e](https://github.com/brinestone/scholaris-api/commit/4331c1e36b274381bad57e19c6e5ef18bc2cd417))
* **app:** cors error ([e9db8fd](https://github.com/brinestone/scholaris-api/commit/e9db8fd9a7e2dd62b9b57b551dfc40330cab1a91))
* **app:** cors error ([3773e26](https://github.com/brinestone/scholaris-api/commit/3773e26534c598a750b1cda0626f336553782763))
* **app:** cors error ([c5ed124](https://github.com/brinestone/scholaris-api/commit/c5ed1240951a27eeb783ab1d1966f8023223fbdf))
* **app:** cors error ([a8acc0e](https://github.com/brinestone/scholaris-api/commit/a8acc0e2e4011f7b3c6c269be5d925a1d30830f7))
* **app:** fix cors 2 ([c802f3d](https://github.com/brinestone/scholaris-api/commit/c802f3d4aaee82f09eb33e7072adf2946c32eedf))
* **app:** updatecors config ([9e7ecd0](https://github.com/brinestone/scholaris-api/commit/9e7ecd02cef9c9c247bccf48744b8aa7a50bbf35))
* **auth,users:** update NewUser endpoint ([4b92635](https://github.com/brinestone/scholaris-api/commit/4b92635a39dc515862b24ec418be1fbcf956d300))
* **auth:** sensitive passwords for sign up model ([5be47e2](https://github.com/brinestone/scholaris-api/commit/5be47e297b5c7de74e4a626dfd96b14cce789f49))
* **forms:** non-optional "ids" field in DeleteFormQuestionGroupsRequest ([b53449f](https://github.com/brinestone/scholaris-api/commit/b53449f2a9aa59580d0b7cae455215d2f11aa76a))
* **forms:** unmocked ListRelations endpoint deploy error ([964bd8f](https://github.com/brinestone/scholaris-api/commit/964bd8f4bb3239517d7c2b9d185cd36be0556e88))
* **institutions:** fix bad query for auto-academic year creation ([ee6d7a8](https://github.com/brinestone/scholaris-api/commit/ee6d7a8f3386a308d51bf32bd7d98340b7e5f5fc))
* **institutions:** fix SQL for migration 6 ([317beec](https://github.com/brinestone/scholaris-api/commit/317beec8324d4f612a95b6285947f23615e2e091))
* **institutions:** fix SQL for migration 6 (2) ([978ec54](https://github.com/brinestone/scholaris-api/commit/978ec5492cae67dec5af266ee998362cd1fa6a2e))
* **institutions:** fix SQL for migration 6 (3) ([99323ee](https://github.com/brinestone/scholaris-api/commit/99323eee871a299fa8f92166c0a672dea7113db3))
* **institutions:** fix SQL for migration 6 (4) ([de07468](https://github.com/brinestone/scholaris-api/commit/de07468ea97312c6a76b0cec17246c60e1388b90))
* **institutions:** fix SQL for migration 6 (5) ([107a344](https://github.com/brinestone/scholaris-api/commit/107a344db72ea8b2862533147a6b51322ef80ed3))
* **instiuttions:** revise api schemas & migrations ([593a728](https://github.com/brinestone/scholaris-api/commit/593a72809b2f8610269bca9827caf7e1b4053bee))
* minor bug fixes ([66d8fbe](https://github.com/brinestone/scholaris-api/commit/66d8fbe7a5398e7248ae3259a91085aaa1369065))
* **permissions:** specify correct credentials method for cloud ([049a829](https://github.com/brinestone/scholaris-api/commit/049a8292c3d565623ea3762792d68a2d4e101eaf))
* **settings:** fix test error on deploy ([124f096](https://github.com/brinestone/scholaris-api/commit/124f09699d748954a1d9fa1f6f6a8ff0ecc0b139))
* **settings:** fix upsert query ([80387b6](https://github.com/brinestone/scholaris-api/commit/80387b6571e2b9dc1b2e27fe9b4a763ccbb7f7e2))
* **settings:** separate endpoints for internal and public setting value updates ([15b0159](https://github.com/brinestone/scholaris-api/commit/15b015979e5031205b27e7542ca3827fc6c416dd))
* **tenants:** 500 error status on FindSubscriptionPlans endpoint ([7e719e2](https://github.com/brinestone/scholaris-api/commit/7e719e2a1b58e59434e3a4b10c1d004f5272447d))
* **tenants:** remove subscribedOnly query parameter from FindTenants endpoint ([534ed04](https://github.com/brinestone/scholaris-api/commit/534ed0471077bd32cb8da648288ead2cd8e113cd))

## 1.1.0 (2024-11-05)


### Features

* **forms:** creation of form questions ([1fa155a](https://github.com/brinestone/scholaris-api/commit/1fa155a1c96cf078825fa0c6cbafa45b44a84d6d))
* **forms:** CRUD endpoints implemented ([4867866](https://github.com/brinestone/scholaris-api/commit/486786693d6e3755e1fba2fa6595060f773b85a9))
* **forms:** deletion of form groups ([ebad11c](https://github.com/brinestone/scholaris-api/commit/ebad11c0acbb292ecba563c40cacbf7b43786d9e))
* **forms:** fetch form info ([8de36d6](https://github.com/brinestone/scholaris-api/commit/8de36d616435e03edd572d1ffecdeb2e8538f76b))
* **forms:** form question update support ([683372b](https://github.com/brinestone/scholaris-api/commit/683372b49a94d92b777409d77aa02559a71f8fb0))
* **forms:** form status toggling ([7f2854c](https://github.com/brinestone/scholaris-api/commit/7f2854c16033acc242920e7d24f73f4bc9a96225))
* **forms:** question creation ([554175f](https://github.com/brinestone/scholaris-api/commit/554175fffec08cb2bca5db0a9985b3d87bbac890))
* **institution:** enrollment creation ([75df58b](https://github.com/brinestone/scholaris-api/commit/75df58bbb66e185f56a5134aa876ddb25a75b6ae))
* **institutions:** create new, lookup and get info ([6f8c5bf](https://github.com/brinestone/scholaris-api/commit/6f8c5bf91edbe7ce686902583d088882a5d90291))
* new form support ([58027bb](https://github.com/brinestone/scholaris-api/commit/58027bb8a90c427a878bc181e28b70d71ea06e30))


### Bug Fixes

* **institutions:** fix SQL for migration 6 ([317beec](https://github.com/brinestone/scholaris-api/commit/317beec8324d4f612a95b6285947f23615e2e091))
* **institutions:** fix SQL for migration 6 (2) ([978ec54](https://github.com/brinestone/scholaris-api/commit/978ec5492cae67dec5af266ee998362cd1fa6a2e))
* **institutions:** fix SQL for migration 6 (3) ([99323ee](https://github.com/brinestone/scholaris-api/commit/99323eee871a299fa8f92166c0a672dea7113db3))
* **institutions:** fix SQL for migration 6 (4) ([de07468](https://github.com/brinestone/scholaris-api/commit/de07468ea97312c6a76b0cec17246c60e1388b90))
* **institutions:** fix SQL for migration 6 (5) ([107a344](https://github.com/brinestone/scholaris-api/commit/107a344db72ea8b2862533147a6b51322ef80ed3))
* **instiuttions:** revise api schemas & migrations ([593a728](https://github.com/brinestone/scholaris-api/commit/593a72809b2f8610269bca9827caf7e1b4053bee))
