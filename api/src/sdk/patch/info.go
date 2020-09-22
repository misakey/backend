package patch

// Info contains fields to patch and model with patched data.
// See Map for more information.
type Info struct {
	Input  []string
	Output []string
	Model  interface{}
}

// Whitelist fields given in parameters from the Info Input member
// it also removes elements saved at same position in Output slice.
func (i *Info) Whitelist(whitelistFields []string) {
	filteredInput := []string{}
	filteredOutput := []string{}
	for pos, field := range i.Input {
		found := false
		for _, whitelistField := range whitelistFields {
			if field == whitelistField {
				found = true
				break
			}
		}
		if found {
			filteredInput = append(filteredInput, field)
			filteredOutput = append(filteredOutput, i.Output[pos])
		}
	}
	i.Input = filteredInput
	i.Output = filteredOutput
}
