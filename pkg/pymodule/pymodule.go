package pymodule

/*
#cgo CXXFLAGS: -std=c++23 -I../../submodules/pybind11/include -I../../submodules/nlohmann_json/include
#cgo pkg-config: python3-embed
*/
import "C"

import (
	_ "github.com/MythicAgents/forgescript/pkg/pymodule/gobindings"
)
