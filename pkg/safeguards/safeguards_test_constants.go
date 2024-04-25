package safeguards

import "fmt"

type TestManifest struct {
	Name         string
	SuccessPaths []string
	ErrorPaths   []string
}

const testManifestDirectory = "tests"

var testDeployments = []TestManifest{
	testManifest_CAI,
	testManifest_CEP,
	testManifest_CL,
	testManifest_CRIP,
	testManifest_DBPDB,
	testManifest_PEA,
	testManifest_RT,
	testManifest_USS,
	testManifest_all,
}

var testError_CAI_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CAI, "CAI-error-manifest.yaml")

var testSuccess_CAI_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CAI, "CAI-success-manifest.yaml")

var testManifest_CAI = TestManifest{
	Name:         Constraint_CAI,
	SuccessPaths: []string{testSuccess_CAI_Standard},
	ErrorPaths:   []string{testError_CAI_Standard},
}

var testError_CEP_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CEP, "CEP-error-manifest.yaml")

var testSuccess_CEP_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CEP, "CEP-success-manifest.yaml")

var testManifest_CEP = TestManifest{
	Name:         Constraint_CEP,
	SuccessPaths: []string{testSuccess_CEP_Standard},
	ErrorPaths:   []string{testError_CEP_Standard},
}

var testError_CRL_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CRL, "CRL-error-manifest.yaml")

var testSuccess_CRL_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CRL, "CRL-success-manifest.yaml")

var testManifest_CL = TestManifest{
	Name:         Constraint_CRL,
	SuccessPaths: []string{testSuccess_CRL_Standard},
	ErrorPaths:   []string{testError_CRL_Standard},
}

var testError_CRIP_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CRIP, "CRIP-error-manifest.yaml")

var testSuccess_CRIP_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_CRIP, "CRIP-success-manifest.yaml")

var testManifest_CRIP = TestManifest{
	Name:         Constraint_CRIP,
	SuccessPaths: []string{testSuccess_CRIP_Standard},
	ErrorPaths:   []string{testError_CRIP_Standard},
}

var testError_DBPDB_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_DBPDB, "DBPDB-error-manifest.yaml")

var testSuccess_DBPDB_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_DBPDB, "DBPDB-success-manifest.yaml")

var testManifest_DBPDB = TestManifest{
	Name:         Constraint_DBPDB,
	SuccessPaths: []string{testSuccess_DBPDB_Standard},
	ErrorPaths:   []string{testError_DBPDB_Standard},
}

var testError_PEA_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_PEA, "PEA-error-manifest.yaml")

var testSuccess_PEA_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_PEA, "PEA-success-manifest.yaml")

var testManifest_PEA = TestManifest{
	Name:         Constraint_PEA,
	SuccessPaths: []string{testSuccess_PEA_Standard},
	ErrorPaths:   []string{testError_PEA_Standard},
}

var testError_RT_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_RT, "RT-error-manifest.yaml")

var testSuccess_RT_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_RT, "RT-success-manifest.yaml")

var testManifest_RT = TestManifest{
	Name:         Constraint_RT,
	SuccessPaths: []string{testSuccess_RT_Standard},
	ErrorPaths:   []string{testError_RT_Standard},
}

var testError_USS_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_USS, "USS-error-manifest.yaml")

var testSuccess_USS_Standard = fmt.Sprintf("%s/%s/%s", testManifestDirectory, Constraint_USS, "USS-success-manifest.yaml")

var testManifest_USS = TestManifest{
	Name:         Constraint_USS,
	SuccessPaths: []string{testSuccess_USS_Standard},
	ErrorPaths:   []string{testError_USS_Standard},
}

var testError_all_Standard_1 = fmt.Sprintf("%s/%s/%s/%s", testManifestDirectory, Constraint_all, "error", "all-error-manifest-1.yaml")
var testError_all_Standard_2 = fmt.Sprintf("%s/%s/%s/%s", testManifestDirectory, Constraint_all, "error", "all-error-manifest-2.yaml")
var testSuccess_all_Standard_1 = fmt.Sprintf("%s/%s/%s/%s", testManifestDirectory, Constraint_all, "success", "all-success-manifest-1.yaml")
var testSuccess_all_Standard_2 = fmt.Sprintf("%s/%s/%s/%s", testManifestDirectory, Constraint_all, "success", "all-success-manifest-2.yaml")

var testManifest_all = TestManifest{
	Name:         "all",
	SuccessPaths: []string{testSuccess_all_Standard_1, testSuccess_all_Standard_2},
	ErrorPaths:   []string{testError_all_Standard_1, testError_all_Standard_2},
}
