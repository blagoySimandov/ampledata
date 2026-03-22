PRODUCTION_DIR := ../internal/services/prompts

PROMPTS := decision-maker extraction

# Current versions
decision-maker_VERSION := v2-clear
extraction_VERSION := v1

define PROMPT_RULES
promote-$(1):
	cp prompts/$(1)-$($(1)_VERSION).txt $(PRODUCTION_DIR)/$(1).txt
	@echo "Promoted $(1)-$($(1)_VERSION).txt -> $(PRODUCTION_DIR)/$(1).txt"

eval-$(1):
	promptfoo eval -c $(1).yaml
endef

$(foreach p,$(PROMPTS),$(eval $(call PROMPT_RULES,$(p))))

.PHONY: $(addprefix promote-,$(PROMPTS)) $(addprefix eval-,$(PROMPTS))
