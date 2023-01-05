package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var outdir = flag.String("outdir", "", "target directory into which to copy -cover tools")
var vflag = flag.Int("v", 0, "verbose trace/debug output level")
var origargs []string

func verbose(vlevel int, s string, a ...interface{}) {
	if *vflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

func runcmd2(tool string, fixup func(c *exec.Cmd), args ...string) string {
	cmd := exec.Command(tool, args...)
	fixup(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", string(out))
		log.Fatalf("%s %+v failed: %s", tool, args, err)
	}
	return string(out)
}

func runcmd(tool string, args ...string) string {
	return runcmd2(tool, func(c *exec.Cmd) {}, args...)
}

func copyFile(srcpath, dstpath string) {
	runcmd("cp", srcpath, dstpath)
}

func writetoolexecsrc(outpath string) {
	f, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(f, "// Auto-generated as part of this run:\n")
	fmt.Fprintf(f, "// %+v\n\n", origargs)
	fmt.Fprintf(f, "%s\n",
		`
package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	tool := os.Args[2]
	tooldir := os.Args[1]
	name := filepath.Base(tool)
	path := filepath.Join(tooldir, name)
	cmd := exec.Command(path, os.Args[3:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

`)
	if err := f.Close(); err != nil {
		log.Fatalf("closing %s: %v", outpath, err)
	}
	verbose(1, "wrote toolexec src to %s", outpath)
}

type tool struct {
	name    string
	srcpath string
	dstpath string
	istool  bool
}

func perform() {

	// Run toolstash -save
	out := runcmd("toolstash", "-v", "save")

	// Interpret the output
	tools := []tool{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		slots := strings.Fields(line)
		if len(slots) != 3 {
			log.Fatalf("unexpected output line from 'toolstash -v save': %s", line)
		}
		x := slots[1]
		t := tool{
			srcpath: x,
			dstpath: slots[2],
			name:    filepath.Base(x),
			istool:  strings.Contains(x, "/tool/"),
		}
		tools = append(tools, t)
	}
	verbose(1, "len(tools) is %d", len(tools))

	goroot := strings.TrimSpace(runcmd("go", "env", "GOROOT"))
	verbose(1, "goroot is %s", goroot)

	// Do the "go install -cover" run.
	setdir := func(c *exec.Cmd) { c.Dir = filepath.Join(goroot, "src") }
	runcmd2("go", setdir, "install", "-coverpkg=all", "cmd")

	// Copy the results to the outdir, and in the process do a
	// toolstash restore,
	for _, t := range tools {
		copyFile(t.srcpath, filepath.Join(*outdir, t.name))
		copyFile(t.dstpath, t.srcpath)
	}

	// Write toolexec wrapper src.
	wrapperpath := filepath.Join(*outdir, "toolexec.go")
	writetoolexecsrc(wrapperpath)

	// Build it.
	setodir := func(c *exec.Cmd) { c.Dir = *outdir }
	runcmd2("go", setodir, "build", "toolexec.go")
}

func main() {
	origargs = make([]string, len(os.Args))
	copy(origargs, os.Args)
	log.SetFlags(0)
	log.SetPrefix("build-cover-tooldir: ")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatalf("unknown extra args")
	}
	if *outdir == "" {
		log.Fatalf("please supply -outdir flag with target directory")
	}
	dirEnts, dirErr := ioutil.ReadDir(*outdir)
	if dirErr != nil {
		log.Fatalf("unable to open -outdir arg %s: %v", *outdir, dirErr)
	}
	if len(dirEnts) != 0 {
		for k, d := range dirEnts {
			fmt.Fprintf(os.Stderr, "%d %+v\n", k, d)
		}
		log.Fatalf("-outdir arg %s appears to not be empty", *outdir)

	}
	verbose(1, "starting perform")
	perform()
	verbose(1, "perform complete")
}
