.PHONY: proto-gen proto-check test

proto-gen:
	./scripts/gen-proto.sh
	rm -rf pb && mv pb_tmp pb

proto-check:
	./scripts/gen-proto.sh
	@if [ -d pb ]; then \
		diff -ru pb pb_tmp || (echo "pb/ is stale; run make proto-gen and commit pb/" && exit 1); \
	else \
		echo "pb/ is missing; run make proto-gen and commit generated files" && exit 1; \
	fi

test: proto-check
	go test ./... -v