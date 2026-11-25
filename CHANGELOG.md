# Changelog

## [0.2.4](https://github.com/hairyhenderson/tplink-ddm-exporter/compare/v0.2.3...v0.2.4) (2026-11-25)


### Dependencies

* build with Go 1.27.4 via a toolchain directive ([#85](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/85)) ([cfce24b](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/cfce24bb8962aeac603354d52e24f9654a2616e7))
* **dockerfile:** update alpine docker tag to v3.24 ([#84](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/84)) ([75e4b2b](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/75e4b2b5db26500fd8a20d34b3a1bcc9d631e538))
* **dockerfile:** update docker/dockerfile docker tag to v1.28.0 ([#71](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/71)) ([094e012](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/094e012399979234b74cfd1ebeb54ba6af806788))
* **dockerfile:** update docker/dockerfile docker tag to v1.29.0 ([#88](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/88)) ([9ef2300](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/9ef2300d2f235cf745155fbd68a020d4dabb6a57))
* **dockerfile:** update golang:1.27.4-alpine docker digest to 70bd7dd ([#86](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/86)) ([16edf2e](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/16edf2e2cc9f30d869a4707a2f3b28f36a32cc47))
* **github-actions:** pin dependencies ([#70](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/70)) ([c05feda](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/c05feda023a4a12e0baaca77669cdb6a5af31d35))
* **github-actions:** update actions/checkout action to v7 ([#78](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/78)) ([d65cf17](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/d65cf17880578b7ae054f4db28d9ab3face0c284))
* **github-actions:** update actions/create-github-app-token action to v3.3.0 ([#72](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/72)) ([359f7f2](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/359f7f2803543616e033460ab667b1b4e7c06268))
* **github-actions:** update actions/create-github-app-token action to v3.4.0 ([#90](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/90)) ([23b773d](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/23b773dd8574d34f8ac981b96631bd447735dbc8))
* **github-actions:** update actions/labeler action to v7 ([#79](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/79)) ([545a1dd](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/545a1dd51404973b6a2287d109a88fc839019002))
* **github-actions:** update actions/setup-go action to v7 ([#80](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/80)) ([e02d26b](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/e02d26bfe818ee2dc45fce6ddc65d9280225cd9d))
* **github-actions:** update actions/setup-go digest to d8cbb80 ([#87](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/87)) ([b0e8f4b](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/b0e8f4bbbfbbe030b47b78b4d2029d90418a83c3))
* **github-actions:** update codecov/codecov-action action to v7 ([#81](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/81)) ([fd9e23a](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/fd9e23a72eefcef86708a60b99fd3cf25d23ab29))
* **github-actions:** update googleapis/release-please-action action to v5 ([#82](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/82)) ([22e8ea4](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/22e8ea4e2c41a46561c53abcbf4bfc77fc022d4a))
* **github-actions:** update registry docker tag to v3 ([#89](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/89)) ([5ea7db6](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/5ea7db6b91399888c5b85030f3b5b7421ed9b8e8))
* **github-actions:** update webiny/action-conventional-commits action to v1.4.2 ([#73](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/73)) ([27d8079](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/27d80798e7e506577816bba67a9552501088e458))
* **gomod:** update module github.com/gosnmp/gosnmp to v1.47.0 ([#74](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/74)) ([0a26217](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/0a262178da2db17fad866154a5919503c9cda651))
* **gomod:** update module github.com/prometheus/client_golang to v1.25.0 ([#75](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/75)) ([9c80015](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/9c800157ea49d21aed01047d951e711d682d6c98))
* **gomod:** update module go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp to v0.72.0 ([#76](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/76)) ([2bded0b](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/2bded0be1e70981c6ce31a7b86f68f39c5502cd4))
* **gomod:** update module go.opentelemetry.io/otel to v1.44.0 [security] ([#55](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/55)) ([edf8982](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/edf8982820320a67d7e9768874f15c5b5a37b80d))
* **gomod:** update module go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc to v1.45.0 [security] ([#57](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/57)) ([e7201c7](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/e7201c7dc20be85ec67810b5bca4dd2df2878760))
* **gomod:** update module go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc to v1.47.0 ([#77](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/77)) ([09c29da](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/09c29dac48c230b5c13efdbebe20919adac2b100))
* **gomod:** update module go.opentelemetry.io/otel/sdk to v1.47.0 [security] ([#58](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/58)) ([72b95ad](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/72b95ad1a6e06edee76e46f263b1319f2adf43f5))
* **gomod:** update module golang.org/x/net to v0.62.0 [security] ([#60](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/60)) ([db84fe8](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/db84fe81560d545e5b90e6fe768088fa9287752e))
* **gomod:** update module google.golang.org/grpc to v1.83.1 [security] ([#63](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/63)) ([bb13e31](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/bb13e3192c51427d1c52108fa5070328abd19f8c))
* update Go to 1.27.3 ([#52](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/52)) ([65c4558](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/65c455875e99095bebf873ac9146e042e003e2a9))

## [0.2.3](https://github.com/hairyhenderson/tplink-ddm-exporter/compare/v0.2.2...v0.2.3) (2026-04-25)


### Bug Fixes

* **perf:** single root BulkWalk and OpenTelemetry tracing ([#37](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/37)) ([0963e2b](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/0963e2bbea3e68fbd3cad6d8eee8e164b6c99dca))

## [0.2.2](https://github.com/hairyhenderson/tplink-ddm-exporter/compare/v0.2.1...v0.2.2) (2026-04-25)


### Bug Fixes

* Handle Integer SNMP type in walkOID ([#35](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/35)) ([00d0fc8](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/00d0fc87fd7374f2bdad70170f49dbd6fbb43ed2))

## [0.2.1](https://github.com/hairyhenderson/tplink-ddm-exporter/compare/v0.2.0...v0.2.1) (2026-04-25)


### Bug Fixes

* Make config/threshold SNMP walks non-fatal ([#33](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/33)) ([ad7f484](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/ad7f48471765ddb81f930f8ff890380e9f36efbf))

## [0.2.0](https://github.com/hairyhenderson/tplink-ddm-exporter/compare/v0.1.0...v0.2.0) (2026-04-25)


### Features

* More state/alarm threshold metrics ([#31](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/31)) ([614b3c4](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/614b3c42221a458ca15bba09fb954cc7ca862dfa))


### Dependencies

* **actions:** Bump codecov/codecov-action from 5 to 6 ([#29](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/29)) ([cdaf0f1](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/cdaf0f1684aa75922993cd51d9632885ff1e0bdb))
* **actions:** Bump docker/login-action from 3 to 4 ([#25](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/25)) ([1b71276](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/1b71276a5a7f48458aabb04d6609d9e591982be5))
* **actions:** Bump docker/setup-buildx-action from 3 to 4 ([#26](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/26)) ([2e2d9a1](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/2e2d9a107edcb04b39c75c8404fa9272fd2698c2))
* **actions:** Bump docker/setup-qemu-action from 3 to 4 ([#24](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/24)) ([f1fc6e9](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/f1fc6e90343f05d497f2030ca39141c91b58fc50))
* **actions:** Bump the actions group across 1 directory with 2 updates ([#27](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/27)) ([4880a6c](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/4880a6cc9ce7f4d47908e5c13012674e20998194))
* **actions:** Bump webiny/action-conventional-commits ([#20](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/20)) ([eba2906](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/eba2906dbcd41b2e032312c3cddcf8a3f83606a9))

## [0.1.0](https://github.com/hairyhenderson/tplink-ddm-exporter/compare/v0.0.1...v0.1.0) (2026-01-10)


### Features

* Support scraping multiple targets ([#13](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/13)) ([9af58a4](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/9af58a42fb5e80308cf0c8ff52e9cc6368a73320))


### Dependencies

* **actions:** Bump codecov/codecov-action from 4 to 5 ([#3](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/3)) ([61c6a53](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/61c6a53e5c8fee3f07b3cfae491832288f5098f0))
* **actions:** Bump golangci/golangci-lint-action from 8 to 9 ([#8](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/8)) ([449d3f8](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/449d3f8f2bd8ac3485e485f5831ee7d6ea711609))
* **actions:** Bump the actions group across 1 directory with 3 updates ([#5](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/5)) ([585a776](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/585a7761e8d8781f6c4bdd9dd1c4f777be23a004))
* **actions:** Bump the actions group with 2 updates ([#7](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/7)) ([2ade329](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/2ade3291c78eb31e8fcc82f7aca7dae8f873d772))
* **docker:** Bump alpine from 3.22 to 3.23 ([#9](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/9)) ([4b1dc2c](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/4b1dc2c3c4c275b9de43724ceb2fe45809763a23))
* **go:** Bump github.com/gosnmp/gosnmp from 1.38.0 to 1.42.1 ([#4](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/4)) ([3bbbb02](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/3bbbb0251dec53e7bbac180094b74206ca13ae8f))
* **go:** Bump github.com/gosnmp/gosnmp from 1.42.1 to 1.43.1 ([#12](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/12)) ([d906221](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/d906221f3e72cf249f84b34c9684f1f5f74a18e2))
* **go:** Bump github.com/prometheus/client_golang from 1.20.5 to 1.23.2 ([#2](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/2)) ([a55b9c9](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/a55b9c9eaca88b99c824a5bf3f3f7a6fbe4b5c31))
* **go:** Bump the otel group with 4 updates ([#10](https://github.com/hairyhenderson/tplink-ddm-exporter/issues/10)) ([f53c219](https://github.com/hairyhenderson/tplink-ddm-exporter/commit/f53c219aeea5bf55ae84b474c7c6561d83d175fc))
