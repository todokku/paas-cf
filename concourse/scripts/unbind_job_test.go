package scripts_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("UnbindJob", func() {
	var (
		jobName string
		stdin   *bytes.Buffer
		session *gexec.Session
	)

	BeforeEach(func() {
		jobName = ""
		stdin = new(bytes.Buffer)
		session = nil
	})

	JustBeforeEach(func() {
		command := exec.Command("./unbind_job.rb", jobName)
		command.Stdin = stdin

		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("error handling", func() {
		Context("empty YAML input", func() {
			BeforeEach(func() {
				jobName = ""
				stdin.Write([]byte(""))
			})

			It("exits with non-zero code and error", func() {
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(BeEmpty())
				Expect(session.Err.Contents()).To(Equal([]byte("Unable to parse YAML hash from the input\n")))
			})
		})

		Context("broken YAML input", func() {
			BeforeEach(func() {
				jobName = ""
				stdin.Write([]byte("? :"))
			})

			It("exits with non-zero code and error", func() {
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(BeEmpty())
				Expect(session.Err).To(gbytes.Say(`psych\.rb`))
			})
		})

		Context("jobs key not in YAML input", func() {
			BeforeEach(func() {
				jobName = ""
				stdin.Write([]byte("a: b"))
			})

			It("exits with non-zero code and error", func() {
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(BeEmpty())
				Expect(session.Err.Contents()).To(Equal([]byte("Can't find job definitions in the input\n")))
			})
		})

		Context("jobs key is not an array", func() {
			BeforeEach(func() {
				jobName = ""
				stdin.Write([]byte("jobs: b"))
			})

			It("exits with non-zero code and error", func() {
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(BeEmpty())
				Expect(session.Err.Contents()).To(Equal([]byte("Jobs definition not an array\n")))
			})
		})

		Context("job not found in jobs key", func() {
			BeforeEach(func() {
				jobName = "fake_job"
				stdin.Write([]byte("jobs:\n  - name: a"))
			})

			It("exits with non-zero code and error", func() {
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(BeEmpty())
				Expect(session.Err.Contents()).To(Equal([]byte("Job fake_job not found in the pipeline\n")))
			})
		})
	})

	Describe("pipeline manipulation", func() {
		Context("paas resource in plan", func() {
			BeforeEach(func() {
				jobName = "myjob"
				stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - get: paas
    passed: ['some_previous_job']
`))
			})

			It("removes passed from paas resource", func() {
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - get: paas
`)))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		Context("paas resource within do in plan", func() {
			BeforeEach(func() {
				jobName = "myjob"
				stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - do:
    - get: paas
      passed: ['some_previous_job']
`))
			})

			It("removes passed from paas resource", func() {
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - do:
    - get: paas
`)))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		Context("paas resource within aggregate in plan", func() {
			BeforeEach(func() {
				jobName = "myjob"
				stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - get: paas
      passed: ['some_previous_job']
`))
			})

			It("removes passed from paas resource", func() {
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - get: paas
`)))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		Context("paas resource within in_parallel in plan", func() {
			BeforeEach(func() {
				jobName = "myjob"
				stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - get: paas
      passed: ['some_previous_job']
`))
			})

			It("removes passed from paas resource", func() {
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - get: paas
`)))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		Context("paas resource within do within aggregate in plan", func() {
			BeforeEach(func() {
				jobName = "myjob"
				stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - do:
      - get: paas
        passed: ['some_previous_job']
`))
			})

			It("removes passed from paas resource", func() {
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - do:
      - get: paas
`)))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		Context("paas resource within do within in_parallel in plan", func() {
			BeforeEach(func() {
				jobName = "myjob"
				stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - do:
      - get: paas
        passed: ['some_previous_job']
`))
			})

			It("removes passed from paas resource", func() {
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - in_parallel:
    - do:
      - get: paas
`)))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})
	})

	Context("paas resource nested throughout plan in aggregate", func() {
		BeforeEach(func() {
			jobName = "myjob"
			stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - get: paas
    passed: 'some_job'
  - in_parallel:
    - get: paas
      passed: ['some', 'jobs']
    - do:
      - get: paas
        passed: ['some_previous_job']
`))
		})

		It("removes passed from all paas resources", func() {
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - get: paas
  - in_parallel:
    - get: paas
    - do:
      - get: paas
`)))
			Expect(session.Err.Contents()).To(BeEmpty())
		})
	})

	Context("paas resource nested throughout plan in in_parallel", func() {
		BeforeEach(func() {
			jobName = "myjob"
			stdin.Write([]byte(`---
jobs:
- name: myjob
  plan:
  - get: paas
    passed: 'some_job'
  - in_parallel:
    - get: paas
      passed: ['some', 'jobs']
    - do:
      - get: paas
        passed: ['some_previous_job']
`))
		})

		It("removes passed from all paas resources", func() {
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(Equal([]byte(`---
jobs:
- name: myjob
  plan:
  - get: paas
  - in_parallel:
    - get: paas
    - do:
      - get: paas
`)))
			Expect(session.Err.Contents()).To(BeEmpty())
		})
	})
})
