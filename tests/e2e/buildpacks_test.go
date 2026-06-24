package e2e

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"strings"

	docker "github.com/docker/docker/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Needed if publishing the builder image when testing image extensions
// const registry = "ghcr.io"
// const repository = "swissdatasciencecenter/renku-frontend-buildpacks"
const testBuilder = "selector"
const builderLoc = "../../builders/" + testBuilder
const customPackage = "cowpy"

var _ = Describe("Testing samples", Label("samples"), Ordered, func() {
	var builderImg string
	var client docker.APIClient
	var err error
	var httpClient http.Client

	BeforeAll(func(ctx SpecContext) {
		builderImg = strings.ToLower(fmt.Sprintf("test-builder-image-%s", getULID()))
		Expect(buildBuilder(ctx, filepath.Join(builderLoc, "builder.toml"), builderImg)).To(Succeed())
		client, err = docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
		Expect(err).ToNot(HaveOccurred())
		httpClient = *http.DefaultClient
	})

	BeforeEach(func() {
		jar, err := cookiejar.New(&cookiejar.Options{})
		Expect(err).ToNot(HaveOccurred())
		httpClient.Jar = jar
	})

	AfterAll(func(ctx SpecContext) {
		if builderImg != "" && client != nil {
			log.Println("Cleaning up builder image")
			err = removeImage(ctx, client, builderImg)
			if err != nil {
				log.Println(err)
			}
		}
		if client != nil {
			log.Println("Closing docker client")
			err = client.Close()
			if err != nil {
				log.Println(err)
			}
		}
	})

	DescribeTableSubtree(
		"coding-agent",
		func(source string) {
			var image string
			var container string
			var port int
			BeforeAll(func(ctx SpecContext) {
				image = strings.ToLower(fmt.Sprintf("test-image-%s", getULID()))
				Expect(buildImage(ctx, builderImg, source, image, map[string]string{})).To(Succeed())
				port = getFreePortOrDie()
				envVars := []string{fmt.Sprintf("RENKU_SESSION_PORT=%d", port)}
				ports := map[int]int{port: port}
				container, err = runImage(ctx, client, image, envVars, ports)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterAll(func(ctx SpecContext) {
				if container != "" && client != nil {
					log.Println("Cleaning up container")
					err = removeContainer(ctx, client, container)
					if err != nil {
						log.Println(err)
					}
				}
				if image != "" && client != nil {
					log.Println("Cleaning up image")
					err = removeImage(ctx, client, image)
					if err != nil {
						log.Println(err)
					}
				}
			})

			Context("when the container is running", func() {
				It("pi should exist as a command in the container", func(ctx SpecContext) {
					_, err := execInContainer(ctx, client, container, []string{"launcher", "pi", "--version"})
					Expect(err).ToNot(HaveOccurred())
				})
				It("users should be able to install pi npm packages", func(ctx SpecContext) {
					_, err := execInContainer(ctx, client, container, []string{"launcher", "pi", "install", "npm:pi-adaptive-thinking"})
					Expect(err).ToNot(HaveOccurred())
				})
				It("claude should exist as a command in the container", func(ctx SpecContext) {
					_, err := execInContainer(ctx, client, container, []string{"launcher", "claude", "--version"})
					Expect(err).ToNot(HaveOccurred())
				})
				It("codex should exist as a command in the container", func(ctx SpecContext) {
					_, err := execInContainer(ctx, client, container, []string{"launcher", "codex", "--version"})
					Expect(err).ToNot(HaveOccurred())
				})
			})
		},
		Entry("using coding-agent sample", "../../samples/coding-agent"),
	)
})
