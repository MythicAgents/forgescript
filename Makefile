MAKEFLAGS += -rR

## Variables
GO = go
SWIG = swig
SED = sed
RM = rm -f
CC = ccache gcc
CXX = ccache g++

## Swig
swig_srcs := pkg/python/bindings/bindings.swigcxx
swig_pkgs := python3-embed
swig_inc := submodules/pybind11/include \
	submodules/nlohmann_json/include

## Go bindings
cgo_exporthdrs := pkg/pymodule/gobindings/gobindings.h

override SWIG_CXXFLAGS += $(addprefix -I,$(swig_inc))
override SWIG_CXXFLAGS += $(foreach pkg,$(swig_pkgs),$(shell pkg-config --cflags $(pkg)))

swig_gen = $(swig_srcs:.swigcxx=_swig.go)
swig_wrapgen = $(swig_srcs:.swigcxx=_wrap.cxx)
swig_deps := $(swig_gen:.go=.go.d)

forgescript_deps = $(cgo_exporthdrs)

clean_files += $(swig_gen) $(swig_wrapgen) $(swig_deps)

.PHONY: all
all: forgescript ## Default target

.PHONY: forgescript
forgescript: $(forgescript_deps) ## Build the forgescript binary
	$(GO) build

.PHONY: forgescript-dbg
forgescript-dbg: ## Build a debug version of the forgescript binary
	@$(MAKE) CGO_CFLAGS=-DNDEBUG GOFLAGS='-gcflags=all=-N -gcflags=all=-l' forgescript

.PHONY: swig
swig: $(swig_gen) ## Generate swig go bindings (for development)

.PHONY: exportheaders
exportheaders: $(cgo_exporthdrs) ## Generate cgo export headers

.PHONY: gen
gen: swig exportheaders ## Generate files

.PHONY: clean
clean: ## Remove built artifacts
	$(GO) clean
	$(RM) $(clean_files)

.PHONY: cacheclean
cacheclean: ## Clean go build cache
	$(GO) clean -cache

.PHONY: help
help: ## Print this help output
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(filter-out %.d,$(MAKEFILE_LIST))) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-20s\033[0m %s\n", $$1, $$2}'

%.h: %.go
	$(GO) tool cgo -objdir $@_obj -exportheader $@ $^
	$(RM) -r $@_obj

%_swig.go %_wrap.cxx: %.swigcxx %_swig.go.d
	$(SWIG) -MT $*_swig.go -MMD -MP -MF $*_swig.go.d -go -cgo -intgosize 64 -module $(@F:%_swig.go=%) -o $*_wrap.cxx -outdir $(@D) $(SWIG_CXXFLAGS) $(addprefix -I,$(<D)) -c++ $<
	$(SED) '1s|^|// +build never\n\n|' $*.go > $*_swig.go
	$(SED) -i '1s|^|// +build never\n\n|' $*_wrap.cxx
	$(RM) $*.go

$(swig_deps):
-include $(swig_deps)
