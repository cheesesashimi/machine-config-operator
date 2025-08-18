package e2e_ocl

import (
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type registryTestCase struct {
	name                    string
	pullspec                string
	credPaths               []string
	inspectErrImageNotFound bool
	inspectErrAccessDenied  bool
	deleteErrImageNotFound  bool
	deleteErrAccessDenied   bool
}

func TestImagePrunerTestCasesExisting(t *testing.T) {
	for _, testCase := range getRegistryTestCases() {
		if !strings.Contains(testCase.name, "digest") {
			continue
		}

		t.Logf("%s", spew.Sdump(testCase))

		// redhatCredPath := "/home/zzlotnik/.docker/config.json"

		// ip, k8sSecret, err := setupImagePrunerForTest(redhatCredPath)
		// require.NoError(t, err)

		// _, digest, err := ip.InspectImage(context.TODO(), testCase.pullspec, k8sSecret, &mcfgv1.ControllerConfig{})
		// require.NoError(t, err)

		// digested, err := utils.ParseImagePullspec(testCase.pullspec, reverseDigest(digest.String()))
		// require.NoError(t, err)

		// t.Log(digested)
	}
}

func reverseDigest(digest string) string {
	split := strings.Split(digest, ":")
	split[1] = reverse(split[1])
	return strings.Join(split, ":")
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func getRegistryTestCases() []registryTestCase {
	noCredPath := ""
	dockerCredPath := "/home/zzlotnik/.creds/dockerhub-creds.json"
	quayCredPath := "/home/zzlotnik/.creds/zzlotnik-quay-admin-creds.json"
	redhatCredPath := "/home/zzlotnik/.docker/config.json"

	return []registryTestCase{
		{
			name:                  "Quay.io existing image",
			pullspec:              "quay.io/skopeo/stable:latest",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                   "Quay.io nonexistent image org",
			pullspec:               "quay.io/notrealgoaway/notrealgoaway:notrealgoaway",
			credPaths:              []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                   "Quay.io nonexistent image repo",
			pullspec:               "quay.io/zzlotnik/notrealgoaway:notrealgoaway",
			credPaths:              []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                    "Quay.io nonexistent image tag",
			pullspec:                "quay.io/zzlotnik/testing:notrealgoaway",
			credPaths:               []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Quay.io nonexistent image digest",
			pullspec:                "quay.io/skopeo/stable@sha256:4665e7a95e6908482feae30c780a5952fc96e6244977ade8f04a958ea8c998d5",
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                  "Quay.io existing image no deletion perms",
			pullspec:              "quay.io/zzlotnik/kmod-enic:latest",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                  "RedHat.io existing image",
			pullspec:              "registry.redhat.io/ubi9/ubi:latest",
			credPaths:             []string{redhatCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                   "RedHat.io nonexistent image org",
			pullspec:               "registry.redhat.io/notrealgoaway/ubi:latest",
			credPaths:              []string{redhatCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                   "RedHat.io nonexistent image repo",
			pullspec:               "registry.redhat.io/ubi9/notrealgoaway:latest",
			credPaths:              []string{redhatCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                    "RedHat.io nonexistent image tag",
			pullspec:                "registry.redhat.io/ubi9/ubi:notrealgoaway",
			credPaths:               []string{redhatCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "RedHat.io nonexistent image digest",
			pullspec:                "registry.redhat.io/ubi9/ubi@sha256:60b11a9d3e53b8449528ce6122e0bf45e96c516f06dcfe6db1468a9834921588",
			credPaths:               []string{redhatCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                  "Red Hat Registry existing image",
			pullspec:              "registry.access.redhat.com/ubi9/ubi:latest",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                    "Red Hat Registry nonexistent image org",
			pullspec:                "registry.access.redhat.com/notrealgoaway/ubi:latest",
			credPaths:               []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Red Hat Registry nonexistent image repo",
			pullspec:                "registry.access.redhat.com/ubi9/notrealgoaway:latest",
			credPaths:               []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Red Hat Registry nonexistent image tag",
			pullspec:                "registry.access.redhat.com/ubi9/ubi:notrealgoaway",
			credPaths:               []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Red Hat Registry nonexisent image digest",
			pullspec:                "registry.access.redhat.com/ubi9/ubi@sha256:60b11a9d3e53b8449528ce6122e0bf45e96c516f06dcfe6db1468a9834921588",
			credPaths:               []string{noCredPath, quayCredPath, dockerCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                  "Docker.io existing image no deletion perms",
			pullspec:              "docker.io/library/python:latest",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                   "Docker.io nonexistent image org",
			pullspec:               "docker.io/notrealgoaway/notrealgoaway:notarealtag",
			credPaths:              []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                   "Docker.io nonexistent image repo - not correctly authed",
			pullspec:               "docker.io/cheesesashimi/notrealgoaway:notarealtag",
			credPaths:              []string{noCredPath, quayCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                    "Docker.io nonexistent image repo - correctly authed",
			pullspec:                "docker.io/cheesesashimi/notrealgoaway:notarealtag",
			credPaths:               []string{dockerCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Docker.io nonexistent image tag",
			pullspec:                "docker.io/cheesesashimi/hotdogcart:notarealtag",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Docker.io nonexistent image digest",
			pullspec:                "docker.io/library/python@sha256:992cb3b24167479655119f438ba34ff0789ab8664f2a1e69cd9e8499c9b1f2b3",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                  "Fedora Registry - existing image",
			pullspec:              "registry.fedoraproject.org/fedora:latest",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                    "Fedora Registry - nonexistent image digest",
			pullspec:                "registry.fedoraproject.org/fedora@sha256:a3322527585b87442ad3d9a43a0f9c36702965d4ed50888221f2a1ae4f420a07",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                  "GitHub image registry - existing image",
			pullspec:              "ghcr.io/open-webui/open-webui:latest",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},
		{
			name:                   "GitHub image registry - nonexistent org",
			pullspec:               "ghcr.io/notrealgoaway/notrealgoaway:notarealtag",
			credPaths:              []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},

		{
			name:                   "GitHub image registry - nonexistent repo",
			pullspec:               "ghcr.io/open-webui/notrealgoaway:notarealtag",
			credPaths:              []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},
		{
			name:                    "GitHub image registry - nonexistent tag",
			pullspec:                "ghcr.io/open-webui/open-webui:notarealtag",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrAccessDenied:   true,
		},
		{
			name:                    "GitHub image registry - nonexistent image digest",
			pullspec:                "ghcr.io/open-webui/open-webui@sha256:805366e33231e62299b753674182928bceab1f30b247ba9e0cc8a32dcb8e934a",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrAccessDenied:   true,
		},
		{
			name:                  "Google image registry - existing image",
			pullspec:              "gcr.io/google.com/cloudsdktool/google-cloud-cli:stable",
			credPaths:             []string{noCredPath, dockerCredPath, quayCredPath},
			deleteErrAccessDenied: true,
		},

		{
			name:                   "Google image registry - nonexistent org",
			pullspec:               "gcr.io/google.com/notrealgoaway/notrealgoaway:notarealtag",
			credPaths:              []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrAccessDenied: true,
			deleteErrAccessDenied:  true,
		},

		{
			name:                    "Google image registry - nonexistent repo",
			pullspec:                "gcr.io/google.com/cloudsdktool/notrealgoaway:notarealtag",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Google image registry - nonexistent tag",
			pullspec:                "gcr.io/google.com/cloudsdktool/google-cloud-cli:notarealtag",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
		{
			name:                    "Google image registry - nonexistent image digest",
			pullspec:                "gcr.io/google.com/cloudsdktool/google-cloud-cli@sha256:a10ac71ff4deb9bd972b2bbc8c4e1f54e3b87cb15b2e51cd16e540e5dd3cad6b",
			credPaths:               []string{noCredPath, dockerCredPath, quayCredPath},
			inspectErrImageNotFound: true,
			deleteErrImageNotFound:  true,
		},
	}
}
