package resolve

import "os/exec"

// execLookPath 抽出来便于测试替换。
var execLookPath = exec.LookPath
