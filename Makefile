# The word after the command is the function to work on, e.g. make run ship.
COMMANDS := run check hint doctor need-what
WHAT := $(firstword $(filter-out $(COMMANDS),$(MAKECMDGOALS)))

BELT ?= 0
W ?= 3

FLAGS = -belt $(BELT) -workers $(W)

.PHONY: $(COMMANDS)

## run: watch one function work, e.g. make run ship
run: need-what
	@t=$$(stty -g 2>/dev/null); go run ./cmd/factory $(FLAGS) $(WHAT); [ -z "$$t" ] || stty $$t

## check: test the lesson of a function, e.g. make check ship
check: need-what
	@go run ./internal/tool check $(WHAT)

## hint: a hint for the first failing test, e.g. make hint ship; run it again for the next hint
hint: need-what
	@go run ./internal/tool hint $(WHAT)

## doctor: check that Go and the build work on this laptop
doctor:
	go version
	go build ./...
	go vet ./internal/... ./cmd/...
	@echo "All good. Start with: make run buildall"

# run saves the terminal settings and put them back afterwards:
# the floor reads single keys, and a crash (e.g. concurrent map writes)
# cannot clean up by itself.
need-what:
	@test -n "$(WHAT)" || (echo "Say which function, e.g.: make $(or $(firstword $(MAKECMDGOALS)),run) ship"; exit 2)

# The function name is a goal too. This rule makes it do
# nothing by itself.
ifneq ($(WHAT),)
$(WHAT):
	@:
endif
