package spack

import (
	"fmt"
	"strings"

	"github.com/vsoch/uptodate/parsers"
)

// UpdateBuildArg will update a spack version build arg
// Should be called from docker.go UpdateArg to ensure intiial checks
// ARG uptodate_spack_<package>=<version>
func UpdateBuildArg(values []string) parsers.Update {

	// This is the full argument with =
	arg := values[0]
	update := parsers.Update{}
	fmt.Printf("Found spack build arg prefix %s\n", arg)

	// Split into buildarg name and value
	parts := strings.SplitN(arg, "=", 2)
	name := parts[0]

	// Keep the original for later comparison
	original := strings.Join(values, " ")
	name = strings.Replace(name, "uptodate_spack_", "", 1)

	// Get versions for current spack package
	pkg := GetSpackPackage(name)

	// Should be sorted with newest first
	if len(pkg.Versions) > 0 {

		updated := parts[0] + "=" + pkg.Versions[0].Name

		// Add any comments back
		for _, extra := range values[1:] {
			updated += " " + extra
		}

		// If the updated version is different from the original, update
		if updated != original {
			update = parsers.Update{Original: original, Updated: updated}
		} else {
			fmt.Println("No difference between:", updated, original)
		}
	}
	return update
}
