.PHONY: build 

# Build the project
build:
	@echo "Building..."
	@mkdir -p bin
	go build -o bin/mcp cmd/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Run the application
run: build
	@echo "Running..."
	./bin/mcp
