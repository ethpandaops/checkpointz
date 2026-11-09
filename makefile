build-web:
	@echo "Building web frontend..."
	@npm --prefix ./web ci --ignore-scripts --no-audit --no-fund && npm --prefix ./web run build
