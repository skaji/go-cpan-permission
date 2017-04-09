package permission

type PermissionResult struct {
	Owner         string   `json:"owner"`
	ModuleName    string   `json:"module_name"`
	CoMaintainers []string `json:"co_maintainers"`
}

type PermissionResults []PermissionResult

func (p PermissionResults) Len() int {
	return len(p)
}

func (p PermissionResults) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PermissionResults) Less(i, j int) bool {
	return p[i].ModuleName < p[j].ModuleName
}
