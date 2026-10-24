.PHONY: toolchain cache test race vet build format binding generated licenses policy process artifact offline diff verify

SOURCE_MANIFEST ?=
RUN := ./scripts/sanitized-run

toolchain cache test race vet build format binding generated licenses policy process artifact offline diff:
	$(RUN) ./scripts/verify-candidate.sh $@ "$(SOURCE_MANIFEST)"

verify:
	$(RUN) ./scripts/verify-candidate.sh all "$(SOURCE_MANIFEST)"
