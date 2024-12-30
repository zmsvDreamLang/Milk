package lua

import (
	"os"
	"path/filepath"
	"plugin"
	"runtime"
)

func OpenLib(L *LState) int {
	libTable := L.NewTable()

	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		L.RaiseError("Error getting home directory: %v", err)
		return 0
	}

	// Build .milklib directory path
	milklibDir := filepath.Join(homeDir, ".milklib")

	if _, err := os.Stat(milklibDir); os.IsNotExist(err) {
		err = os.Mkdir(milklibDir, 0755)
		if err != nil {
			L.RaiseError("Error creating .milklib directory: %v", err)
			return 0
		}
	}

	// Read all plugin files in .milklib directory
	files, err := os.ReadDir(milklibDir)
	if err != nil {
		L.RaiseError("Error reading .milklib directory: %v", err)
		return 0
	}

	for _, file := range files {
		if isPluginFile(file.Name()) {
			soPath := filepath.Join(milklibDir, file.Name())

			// Load plugin file
			p, err := plugin.Open(soPath)
			if err != nil {
				L.RaiseError("Error loading plugin %s: %v", soPath, err)
				continue
			}

			// Lookup ExportToMilkFunc function
			exportFuncSymbol, err := p.Lookup("ExportToMilkFunc")
			if err != nil {
				L.RaiseError("ExportToMilkFunc function not found in %s: %v", soPath, err)
				continue
			}

			// Lookup ExportToMilkTbl function
			exportTblSymbol, err := p.Lookup("ExportToMilkTbl")
			if err != nil {
				L.RaiseError("ExportToMilkTbl function not found in %s: %v", soPath, err)
				continue
			}

			ExportToMilkFunc, ok := exportFuncSymbol.(func() map[string]LGFunction)
			if !ok {
				L.RaiseError("ExportToMilkFunc in %s is not of the correct type", soPath)
				continue
			}

			ExportToMilkTblFunc, ok := exportTblSymbol.(func(*LState) *LTable)
			if !ok {
				L.RaiseError("ExportToMilkTbl in %s is not of the correct type", soPath)
				continue
			}

			// Call ExportToMilk function to get exported functions
			exportedFuncs := ExportToMilkFunc()

			// Call ExportToMilkTbl function to get exported table
			exportedTbl := ExportToMilkTblFunc(L)

			// Register exported functions into "lib" table
			for name, fn := range exportedFuncs {
				L.SetField(libTable, name, L.NewFunction(fn))
			}

			// Add entries from the exported table to "lib" table
			exportedTbl.ForEach(func(name LValue, value LValue) {
				L.SetField(libTable, name.String(), value)
			})
		}
	}

	// Set library as global variable "lib"
	L.SetGlobal("lib", libTable)

	// Return library table
	L.Push(libTable)

	return 1
}

// isPluginFile 检查文件是否为插件文件，根据操作系统返回相应的扩展名
func isPluginFile(filename string) bool {
	switch runtime.GOOS {
	case "windows":
		return filepath.Ext(filename) == ".dll"
	case "linux":
		return filepath.Ext(filename) == ".so"
	case "darwin": // macOS
		return filepath.Ext(filename) == ".so"
	default:
		return false
	}
}
