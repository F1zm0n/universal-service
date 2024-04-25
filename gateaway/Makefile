tailwind:
	npx tailwindcss -i ./view/tailwind/input.css -o ./view/static/output.css --watch

run:
	@npx tailwindcss -i ./view/tailwind/input.css -o ./view/static/output.css 
	@templ generate ./view
	@go run main.go

templ:
	@templ generate -watch -proxy=http://localhost:3000

