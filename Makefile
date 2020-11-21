.DEFAULT_GOAL := help


.PHONY: help
help: ## View help text for all available commands
	@grep -E '^.*:.*?## .*$$' $(MAKEFILE_LIST) \
		| grep -v '@grep' | grep -v 'BEGIN' | sort \
		| awk 'BEGIN {FS = ":.*? ## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: demo
demo: ## Render demo spreadsheet
	@docker-compose -f demo/docker-compose.yml run demo
	@echo "Rendered file path: demo/report_result.xlsx"
