#
#  SPDX-License-Identifier: AGPL-3.0-only
#  Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
#

.PHONY: schema-resolver

schema-resolver:
	@unset GOOS && \
    unset GOARCH && \
    unset GOARM && \
    go generate ./...