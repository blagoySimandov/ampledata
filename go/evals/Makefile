PRODUCTION_DIR := ../internal/services/prompts

promote-decision-maker:
	cp prompts/decision-maker-v1.txt $(PRODUCTION_DIR)/decision-maker.txt
	@echo "Promoted decision-maker-v1.txt -> $(PRODUCTION_DIR)/decision-maker.txt"

promote-extraction:
	cp prompts/extraction-v1.txt $(PRODUCTION_DIR)/extraction.txt
	@echo "Promoted extraction-v1.txt -> $(PRODUCTION_DIR)/extraction.txt"

eval-decision-maker:
	promptfoo eval -c decision-maker.yaml

eval-extraction:
	promptfoo eval -c extraction.yaml

.PHONY: promote-decision-maker promote-extraction eval-decision-maker eval-extraction
