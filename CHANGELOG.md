# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [2.2.0](https://github.com/brinestone/scholaris-api/compare/v2.1.0...v2.2.0) (2024-12-12)


### Features

* **blob:** basic upload/download support ([f42c7ed](https://github.com/brinestone/scholaris-api/commit/f42c7edebab74816a9487ad6bf106481d3520ec1))
* **instiutions:** remove params from lookup endpoint ([3cc6467](https://github.com/brinestone/scholaris-api/commit/3cc64674d66e5a2b6e101be8d1bcfb8a2bb2c9bf))
* **permissions:** cache ListRelations endpoint ([a5e7111](https://github.com/brinestone/scholaris-api/commit/a5e7111b08c424137096de20ff7c1bf4c44abd39))
* **permissions:** relations listing api ([79304af](https://github.com/brinestone/scholaris-api/commit/79304af10ec624a1799acb5498b9e3d69a663071))
* **tenants:** name availability endpoint ([df9ae87](https://github.com/brinestone/scholaris-api/commit/df9ae8782cd7a48ea2a7de2e5d415a6cbef0e861))
* **tenants:** remove params for lookup ([95895f0](https://github.com/brinestone/scholaris-api/commit/95895f05cb71b8ca04c7cac2863b932f0bdace68))


### Bug Fixes

* **permissions:** fix bad validation for ListRelationsRequest ([30f5117](https://github.com/brinestone/scholaris-api/commit/30f5117ecee402a36295ad5345e67b23bcf899fa))
* **permissions:** fix bad validation for ListRelationsRequest ([aecd933](https://github.com/brinestone/scholaris-api/commit/aecd93387d3b4f0580f6cb27b445f6b790ab2344))
* **permissions:** fix bad validation for ListRelationsRequest ([27e1b3f](https://github.com/brinestone/scholaris-api/commit/27e1b3f58ea27447b0474e86c04834bf4b2263bd))
* **permissions:** fix bad validation for ListRelationsRequest ([35e5238](https://github.com/brinestone/scholaris-api/commit/35e5238d2ca34aca494cb0533cf4aed8ed134e4d))
* **permissions:** fix list relations endpoin ([ede8ca4](https://github.com/brinestone/scholaris-api/commit/ede8ca4250d336b8eab61532c5bdb6c84df5216d))
* **permissions:** fix listRelations endpoint caching key gen ([ae6f344](https://github.com/brinestone/scholaris-api/commit/ae6f34415c79453e05ae098fa9858c1f287bdfac))
* **permissions:** fix listRelations endpoint caching key gen ([65eecf3](https://github.com/brinestone/scholaris-api/commit/65eecf39a0de4fb23641446966991be30705b669))
* **permissions:** fix listRelations endpoint caching key gen ([a32875f](https://github.com/brinestone/scholaris-api/commit/a32875f5720716cdc009496f994342054264242e))
* **permissions:** fix listRelations endpoint caching key gen 2 ([55f3c3c](https://github.com/brinestone/scholaris-api/commit/55f3c3c338d9528ec6fe12f55241b06626880fbc))
* **permissions:** fix listRelations endpoint caching key gen 3 ([fa293ef](https://github.com/brinestone/scholaris-api/commit/fa293ef6f25ee46759d5adad17083e452b9e4c2b))

## [2.1.0](https://github.com/brinestone/scholaris-api/compare/v2.0.1...v2.1.0) (2024-11-28)


### Features

* **auth,users:** clerk integration ([ae09124](https://github.com/brinestone/scholaris-api/commit/ae09124d3ff16a80e95bbbcc7990d60628ef54b3))
* **auth:** per provider jwt-validation/verification ([f1c4c2a](https://github.com/brinestone/scholaris-api/commit/f1c4c2aeed61f520efd537aedd573ca9b4549dfc))


### Bug Fixes

* **auth:** clerk jwt auth & user creation webhook ([d846710](https://github.com/brinestone/scholaris-api/commit/d84671016b762fe48647bba2e581ba7224b78f97))
* **auth:** clerk jwt auth & user creation webhook 2 ([bb39a3a](https://github.com/brinestone/scholaris-api/commit/bb39a3a6de8bafb21c1c114a978020dae7488ea1))
* **auth:** clerk jwt auth & user creation webhook 3 ([5c97cfd](https://github.com/brinestone/scholaris-api/commit/5c97cfd3adccb1088669465c71fd9f4012ac9dde))
* **auth:** clerk jwt auth & user creation webhook 4 ([ac273b5](https://github.com/brinestone/scholaris-api/commit/ac273b503fb444df887d03cb0bb68f02c3d8e581))
* **auth:** clerk jwt auth & user creation webhook 5 ([0db6861](https://github.com/brinestone/scholaris-api/commit/0db68617a82f677617f55131209f6b7b9d95fe12))
* **auth:** clerk jwt auth & user creation webhook 6 ([0d53231](https://github.com/brinestone/scholaris-api/commit/0d532315f7c3deb850b723c1902b8865da32b0df))
* **institutions:** enfore permissions while retrieving instiutions ([7dedf57](https://github.com/brinestone/scholaris-api/commit/7dedf570ba78800241961b796953e4e801c0094e))
* **institutions:** update academic-year_test.go ([8eba1c7](https://github.com/brinestone/scholaris-api/commit/8eba1c737a20078b848c9c018e46ab14588c3c9d))
* **tenants:** fix sql erroneous quries ([7fa4cc1](https://github.com/brinestone/scholaris-api/commit/7fa4cc102ed25fdc6fb366d9f98ca5fda8de6c68))
* **tenants:** minor refactor ([8ae8cc6](https://github.com/brinestone/scholaris-api/commit/8ae8cc620c34156f3a411c826cff594fae48054b))
* **tenants:** use basic subscription plan on new tenant ([42247fc](https://github.com/brinestone/scholaris-api/commit/42247fc6a0a1c31720d0f28acad30d612da41a27))
* **users:** cache external users on fetch ([3760b75](https://github.com/brinestone/scholaris-api/commit/3760b759aebafa65ca2536366f6b8d93bb61929a))
* **users:** update external user creation DB function ([1b3ed91](https://github.com/brinestone/scholaris-api/commit/1b3ed91a5eba0083a4ad7f679ac1698257286b6a))
* **webhooks,users,auth:** fix clerk user.created integration ([f21f07f](https://github.com/brinestone/scholaris-api/commit/f21f07ff9db963762dea994cc7bc91c30cf625f2))

### [2.0.1](https://github.com/brinestone/scholaris-api/compare/v2.0.0...v2.0.1) (2024-11-20)


### Bug Fixes

* **auth:** fix validate password on account deletion ([d85552e](https://github.com/brinestone/scholaris-api/commit/d85552e4b333347b1b17495a633e05d71dbbcb52))

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
