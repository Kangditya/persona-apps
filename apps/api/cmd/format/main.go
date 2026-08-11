package main

import (
    "bytes"
    "flag"
    "fmt"
    "go/format"
    "go/scanner"
    "go/token"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

const literalIndent = "    "

func main() {
    write := flag.Bool("write", false, "write formatted files")
    check := flag.Bool("check", false, "fail when files need formatting")
    flag.Parse()

    files, err := sourceFiles(flag.Args())
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    changed := 0
    for _, path := range files {
        before, err := os.ReadFile(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
            os.Exit(1)
        }
        after, err := formatted(before)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
            os.Exit(1)
        }
        if bytes.Equal(before, after) {
            continue
        }
        changed++
        fmt.Println(path)
        if *write {
            info, err := os.Stat(path)
            if err != nil {
                fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
                os.Exit(1)
            }
            if err := os.WriteFile(path, after, info.Mode()); err != nil {
                fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
                os.Exit(1)
            }
        }
    }

    fmt.Printf("Go files scanned: %d\nGo files changed: %d\n", len(files), changed)
    if *check && changed != 0 {
        os.Exit(1)
    }
}

func sourceFiles(args []string) ([]string, error) {
    if len(args) != 0 {
        files := make([]string, 0, len(args))
        for _, arg := range args {
            path, err := filepath.Abs(arg)
            if err != nil {
                return nil, err
            }
            files = append(files, path)
        }
        return files, nil
    }

    root, err := repositoryRoot()
    if err != nil {
        return nil, err
    }
    output, err := exec.Command("git", "-C", root, "ls-files", "-z", "--", "*.go").Output()
    if err != nil {
        return nil, fmt.Errorf("list tracked Go files: %w", err)
    }

    var files []string
    for _, raw := range bytes.Split(output, []byte{0}) {
        if len(raw) == 0 {
            continue
        }
        files = append(files, filepath.Join(root, filepath.FromSlash(string(raw))))
    }
    return files, nil
}

func repositoryRoot() (string, error) {
    output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
    if err != nil {
        return "", fmt.Errorf("find repository root: %w", err)
    }
    return strings.TrimSpace(string(output)), nil
}

func formatted(source []byte) ([]byte, error) {
    canonical, err := format.Source(source)
    if err != nil {
        return nil, err
    }
    return normalizeIndent(canonical), nil
}

func normalizeIndent(source []byte) []byte {
    protected := protectedLines(source)
    var output bytes.Buffer
    line := 1
    for offset := 0; offset < len(source); {
        end := bytes.IndexByte(source[offset:], '\n')
        if end == -1 {
            end = len(source)
        } else {
            end += offset
        }

        if protected[line] {
            output.Write(source[offset:end])
        } else {
            for i := offset; i < end; i++ {
                switch source[i] {
                case '\t':
                    output.WriteString(literalIndent)
                case ' ':
                    output.WriteByte(source[i])
                default:
                    output.Write(source[i:end])
                    i = end
                }
            }
        }

        if end < len(source) {
            output.WriteByte('\n')
        }
        offset = end + 1
        line++
    }
    return output.Bytes()
}

func protectedLines(source []byte) map[int]bool {
    file := token.NewFileSet().AddFile("source.go", -1, len(source))
    var scan scanner.Scanner
    scan.Init(file, source, nil, scanner.ScanComments)
    protected := map[int]bool{}
    for {
        position, tokenType, literal := scan.Scan()
        if tokenType == token.EOF {
            return protected
        }
        if tokenType != token.STRING && tokenType != token.COMMENT {
            continue
        }
        if tokenType == token.STRING && (len(literal) == 0 || literal[0] != 96) {
            continue
        }
        if tokenType == token.COMMENT && !strings.HasPrefix(literal, "/*") {
            continue
        }

        first := file.Line(position)
        last := file.Line(position + token.Pos(len(literal)-1))
        for line := first + 1; line < last; line++ {
            protected[line] = true
        }
    }
}
