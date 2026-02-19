## [0.0.6](https://github.com/tduyng/codeme/compare/v0.0.5..v0.0.6) - 2026-02-19

### Features

- Add read-only connection and rebuild-summaries command - ([94f6186](https://github.com/tduyng/codeme/commit/94f61861c735f8c374ffa167c7eb38e054a17aa9))
- Add summary tables and caching layer - ([7955eaf](https://github.com/tduyng/codeme/commit/7955eafc1dd97f30dfc403a5770c5d18aabce486))

### Bug Fixes

- *(core)* Unify query path and fix resource leaks - ([992f56d](https://github.com/tduyng/codeme/commit/992f56d9c53477528ffd36dde35b347f16552c18))
- Correct period filtering for most productive day and highest daily output - ([94407a3](https://github.com/tduyng/codeme/commit/94407a3a8601a314ec601eeb1a7b737e1a90d822))
- Add GetDailySummaries method for future optimization - ([ba4582a](https://github.com/tduyng/codeme/commit/ba4582ae66df16e1276564f9fa9b889098d8b26d))
- Update summary tables in real-time on SaveActivity - ([00878fb](https://github.com/tduyng/codeme/commit/00878fb80bfd10eebd77828fa8e2481b8689111a))

### Refactor

- Remove unused funcs - ([2ce15d2](https://github.com/tduyng/codeme/commit/2ce15d2ae40cd284063cd148eeb2f4cd99378924))

### Performance

- Replace UUID with timestamp-based ID generation - ([c4ea280](https://github.com/tduyng/codeme/commit/c4ea280db817f7c52e2a6cecc47ddc2fae8aec73))
- Merge activity loops and add covering index - ([3e94c23](https://github.com/tduyng/codeme/commit/3e94c23209eb36ba8c7ab8df29253878a8cbe29e))
- Remove redundant sorts and optimize session grouping - ([53e7c58](https://github.com/tduyng/codeme/commit/53e7c588f1b9c3424cb70950c0c3300d3f6cce14))

### Miscellaneous Tasks

- Update README - ([c1ed52b](https://github.com/tduyng/codeme/commit/c1ed52b313f373c2beb8da298661fc57aea61bda))
- Add uninstall command - ([ce9595d](https://github.com/tduyng/codeme/commit/ce9595d5c79f3163e11d1966ff450a70bfa3c7e1))
- Make installation more clear - ([7602af8](https://github.com/tduyng/codeme/commit/7602af8839432ac649de14cb1726de7973c62655))
## [0.0.5](https://github.com/tduyng/codeme/compare/v0.0.4..v0.0.5) - 2026-02-15

### Miscellaneous Tasks

- Change sqlite lib to pur go, do not need gcc enable - ([9f5c6ae](https://github.com/tduyng/codeme/commit/9f5c6ae5ca05819975a05f42ab8ddba9b556cb34))
## [0.0.4](https://github.com/tduyng/codeme/compare/v0.0.3..v0.0.4) - 2026-02-15

### Miscellaneous Tasks

- Add version in archive release file - ([9d69dca](https://github.com/tduyng/codeme/commit/9d69dca8f0fa6f473af59f7a21fca4198b780460))
- Do not need LATEST_CHANGELOG file - ([907faef](https://github.com/tduyng/codeme/commit/907faeffc4f8317d5025e1e27a9d9046ccabad2a))
## [0.0.3](https://github.com/tduyng/codeme/compare/v0.0.2..v0.0.3) - 2026-02-15

### Bug Fixes

- Calculate correct first date of the week - ([cc2bb2e](https://github.com/tduyng/codeme/commit/cc2bb2eaa6252cb7694c09879654c90e7f9ee332))

### Miscellaneous Tasks

- Avoid dirty issues when release with goreleaser - ([3f5631f](https://github.com/tduyng/codeme/commit/3f5631ffc6f475d4514411fd85e38b29a8ba8b65))
- Show better release changelog - ([4008770](https://github.com/tduyng/codeme/commit/40087707ce078983f6e4e164fddd43e85d92ab83))
## [0.0.2](https://github.com/tduyng/codeme/compare/v0.0.1..v0.0.2) - 2026-02-01

### Features

- *(stats)* Add more details about achievements - ([296df63](https://github.com/tduyng/codeme/commit/296df63de9a34172e7c8edcfb57eb41ebc4aef83))
- Simplify architectures - ([b33bc2b](https://github.com/tduyng/codeme/commit/b33bc2bff37c3b849202e43fa541799638429122))
- Redesign typing for more flexible - ([c613cdd](https://github.com/tduyng/codeme/commit/c613cdd1b0eb0f250c338b366038e8f72e2f5d24))
- Add weekly/monthly stats, heatmap, sessions, and achievements - ([fde4676](https://github.com/tduyng/codeme/commit/fde4676a6a67132ebe9be509872eab9f6cac4185))

### Bug Fixes

- *(tests)* Isolate test databases to prevent production data loss - ([080259c](https://github.com/tduyng/codeme/commit/080259c7e3e1f172ddcc604d57d1f91d070d3c13))
- Display todayHourly correctly - ([67e226d](https://github.com/tduyng/codeme/commit/67e226d37c26c8276d14c20f21573a99b7f040df))
- Correct today command stats calculation and timestamp storage - ([e8225f7](https://github.com/tduyng/codeme/commit/e8225f76dde03c3d28646bad041b1862225f0afa))
- Use git describe for version format - ([eb3964a](https://github.com/tduyng/codeme/commit/eb3964a6df0c085b218288d8f2f9aec7fed9b07b))
## [0.0.1] - 2026-01-12

### Features

- Init codeme server with tracking and stats - ([d25a1ae](https://github.com/tduyng/codeme/commit/d25a1aedce9c0620fb17432243f790685b2edfa1))
