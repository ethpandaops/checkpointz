build-web:
	@echo "Building web frontend..."
	@npm --prefix ./web install && npm --prefix ./web run build
