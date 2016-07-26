/*
The MIT License (MIT)
Copyright (c) 2016 Lau, Chok Sheak (for software "git-sanity-checker")
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/kardianos/osext"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

/**************************************************************************/

// Types.

type ruleCheckArgs struct {
	filePath     string
	fileAsString string
	fileAsLines  []string
}

type checkFuncType func(args ruleCheckArgs) string

type rule struct {
	name           string
	argument       string
	fileTypeFlags  int
	fileExtensions []string
	checkFunc      checkFuncType
}

/**************************************************************************/

// Constants.

const (
	configFileName = "git-sanity-checker.cfg"
	flagTextFile   = 1
	flagBinaryFile = 2
)

/**************************************************************************/

// Variables.

var (
	rulesDefinitions               = initRulesDefinitions()
	csharpKeywordsWithSpacedParens = []string{
		"catch",
		"for",
		"foreach",
		"if",
		"lock",
		"switch",
		"using",
		"while",
	}
)

/**************************************************************************/

// Utilities.

func selectString(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}

func pluralS(count int) string {
	return selectString(count == 1, "", "s")
}

func fatal(s string) {
	fmt.Println(s)
	os.Exit(1)
}

func getScriptDirectory() string {
	dir, err := osext.ExecutableFolder()
	if err != nil {
		fatal(err.Error())
	}
	return dir
}

func readFileString(filePath string) string {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fatal("Cannot read file " + filePath + ": " + err.Error())
	}
	return string(bytes)
}

func readFileLines(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		fatal("Cannot read file \"" + filePath + "\": " + err.Error())
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func execAndGetOutput(command string, args []string) string {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fatal("Cannot execute command \"" + command + "\": " + string(output))
	}
	return string(output)
}

func convertStringToLines(s string, returnNonEmptyOnly bool) []string {
	if s == "" {
		return []string{}
	}
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.Replace(s, "\r", "\n", -1)
	lines := strings.Split(s, "\n")

	if !returnNonEmptyOnly {
		return lines
	}

	var nonEmptyLines []string
	for _, line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return nonEmptyLines
}

func readFirstNChars(filePath string, bytesCount int) string {
	file, err := os.Open(filePath)
	if err != nil {
		fatal("Cannot read file \"" + filePath + "\": " + err.Error())
	}
	defer file.Close()

	bytes := make([]byte, bytesCount)
	file.Read(bytes)
	return string(bytes)
}

func isControlCharacter(char rune) bool {
	return (char == 127) || (((0 <= char) && (char <= 31)) && (char != 9) && (char != 10) && (char != 13))
}

func hasControlCharacters(line string) bool {
	for _, char := range line {
		if isControlCharacter(char) {
			return true
		}
	}
	return false
}

func getCanonicalPath(filePath string) string {
	absFile, err := filepath.Abs(filePath)
	if err == nil {
		filePath = absFile
	}
	filePath = filepath.Clean(filePath)
	return filePath
}

func stringArrayContains(array []string, s string) bool {
	for _, element := range array {
		if s == element {
			return true
		}
	}
	return false
}

func stringArrayReverse(array []string) {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
}

/**************************************************************************/

// Config.

func loadConfigFile() []rule {
	cwd := getScriptDirectory()
	configFilePath := path.Join(cwd, configFileName)
	lines := readFileLines(configFilePath)
	rules := parseConfigRules(configFilePath, lines)
	return rules
}

func parseConfigRules(configFilePath string, lines []string) []rule {
	var rules []rule
	for _, line := range lines {
		line = strings.Trim(line, " \t")
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		r := parseRule(line)
		rules = append(rules, r)
	}

	if len(rules) == 0 {
		fatal("No rules are loaded from \"" + configFilePath + "\"")
	}

	return rules
}

/**************************************************************************/

// Rules.

func initRulesDefinitions() map[string]rule {
	defs := make(map[string]rule)

	addRule(defs, "DoNothing", ruleCheckDoNothing, flagTextFile, "")
	addRule(defs, "NoTabs", ruleCheckNoTabs, flagTextFile, "")
	addRule(defs, "NoLeadingSpaces", ruleCheckNoLeadingSpaces, flagTextFile, "")
	addRule(defs, "TabsVsSpacesOnly", ruleCheckTabsVsSpacesOnly, flagTextFile, "")
	addRule(defs, "ConsistentNewlines", ruleCheckConsistentNewlines, flagTextFile, "")
	addRule(defs, "ConsistentIndentWidth", ruleCheckConsistentIndentWidth, flagTextFile, "")
	addRule(defs, "BadNameSpace", ruleCheckBadNameSpace, flagTextFile, ".cs")
	addRule(defs, "BadClassName", ruleCheckBadClassName, flagTextFile, ".cs")
	addRule(defs, "NoMultiplePublicClasses", ruleCheckNoMultiplePublicClasses, flagTextFile, ".cs")
	addRule(defs, "WindowsNewlines", ruleCheckWindowsNewlines, flagTextFile, "")
	addRule(defs, "LinuxNewlines", ruleCheckLinuxNewlines, flagTextFile, "")
	addRule(defs, "OldMacNewlines", ruleCheckOldMacNewlines, flagTextFile, "")
	addRule(defs, "NeedSpaceAfterKeyword", ruleCheckNeedSpaceAfterKeyword, flagTextFile, ".cs")

	return defs
}

// fileTypeFlags is a bit map int like flagTextFile | flagBinaryFile.
// fileExtensions is a pipe separated string like ".cs|.java". Need the dot also.
func addRule(rulesMap map[string]rule, ruleName string, checkFunc checkFuncType, fileTypeFlags int, fileExtensions string) {
	fileExtensionsArray := []string{}
	if fileExtensions != "" {
		fileExtensionsArray = strings.Split(fileExtensions, "|")
	}

	rulesMap[ruleName] = rule{
		name:           ruleName,
		checkFunc:      checkFunc,
		fileTypeFlags:  fileTypeFlags,
		fileExtensions: fileExtensionsArray,
	}
}

func getRuleMustExist(ruleName string) rule {
	r, ok := rulesDefinitions[ruleName]
	if !ok {
		fatal("Unrecognized rule: " + ruleName)
	}
	return r
}

func parseRule(line string) rule {
	spaceIndex := strings.Index(line, " ")
	if spaceIndex < 0 {
		// No need to clone rules that have no arguments.
		return getRuleMustExist(line)
	}

	// Need to clone rules with arguments.
	name := line[0:spaceIndex]
	ruleDef := getRuleMustExist(name)
	newRule := cloneRule(ruleDef)
	newRule.argument = line[spaceIndex+1:]
	return newRule
}

func cloneRule(r rule) rule {
	return rule{
		name:           r.name,
		argument:       r.argument,
		checkFunc:      r.checkFunc,
		fileExtensions: r.fileExtensions,
	}
}

func fileError(filePath, err string) string {
	return filePath + ": " + err
}

func fileAndLineError(filePath string, lineNum int, err string) string {
	return filePath + ":" + strconv.Itoa(lineNum) + ": " + err
}

func checkEachLine(filePath string, lines []string, errMsg string, isBad func(string) bool) string {
	for lineNum, line := range lines {
		if isBad(line) {
			return fileAndLineError(filePath, lineNum+1, errMsg)
		}
	}
	return ""
}

func ruleCheckDoNothing(args ruleCheckArgs) string {
	return ""
}

func ruleCheckNoTabs(args ruleCheckArgs) string {
	return checkEachLine(args.filePath, args.fileAsLines, "Tabs not allowed", lineContainsTab)
}

func lineContainsTab(line string) bool {
	return strings.ContainsRune(line, '\t')
}

func ruleCheckNoLeadingSpaces(args ruleCheckArgs) string {
	return checkEachLine(args.filePath, args.fileAsLines, "Leading spaces not allowed", lineHasLeadingSpaces)
}

func lineHasLeadingSpaces(line string) bool {
	if line == "" {
		return false
	}
	hasSpace := false
	for _, r := range line {
		if r == ' ' {
			hasSpace = true
		} else {
			return hasSpace
		}
	}
	return hasSpace
}

func ruleCheckTabsVsSpacesOnly(args ruleCheckArgs) string {
	hasTabIndent := false
	hasSpaceIndent := false

	for lineNum, line := range args.fileAsLines {
		for _, r := range line {
			if r == ' ' {
				if hasTabIndent {
					return fileAndLineError(args.filePath, lineNum+1, "Found first space indent with prior tab indents")
				}
				hasSpaceIndent = true
			} else if r == '\t' {
				if hasSpaceIndent {
					return fileAndLineError(args.filePath, lineNum+1, "Found first tab indent with prior space indents")
				}
				hasTabIndent = true
			} else {
				break
			}
		}
	}

	return ""
}

func ruleCheckConsistentNewlines(args ruleCheckArgs) string {
	hasWindows := strings.Contains(args.fileAsString, "\r\n")
	s := strings.Replace(args.fileAsString, "\r\n", "", -1)

	hasLinux := strings.ContainsRune(s, '\n')
	hasOldMac := strings.ContainsRune(s, '\r')

	count := 0
	if hasWindows {
		count++
	}
	if hasLinux {
		count++
	}
	if hasOldMac {
		count++
	}

	if count > 1 {
		return fileError(args.filePath, "File uses inconsistent newlines")
	}
	return ""
}

func ruleCheckConsistentIndentWidth(args ruleCheckArgs) string {
	all3Spaces := true
	all4Spaces := true
	firstNon3SpaceLineNum := 0
	firstNon4SpaceLineNum := 0

	for lineNum, line := range args.fileAsLines {
		numSpaces := 0
		for _, r := range line {
			if r != ' ' {
				break
			}
			numSpaces++
		}

		if all3Spaces && (numSpaces%3) != 0 {
			all3Spaces = false
			firstNon3SpaceLineNum = lineNum + 1
		}
		if all4Spaces && (numSpaces%4) != 0 {
			all4Spaces = false
			firstNon4SpaceLineNum = lineNum + 1
		}
	}

	if !all3Spaces && !all4Spaces {
		if !all4Spaces {
			return fileAndLineError(args.filePath, firstNon4SpaceLineNum, "File has first non 4-space indent")
		}
		return fileAndLineError(args.filePath, firstNon3SpaceLineNum, "File has first non 3-space indent")
	}
	return ""
}

func ruleCheckBadNameSpace(args ruleCheckArgs) string {
	pathAsNameSpace := ""

	for lineNum, line := range args.fileAsLines {
		line = strings.Trim(line, " \t")
		if !strings.HasPrefix(line, "namespace ") {
			continue
		}

		namespace := line[len("namespace "):]

		if pathAsNameSpace == "" {
			pathAsNameSpace = getFilePathAsNameSpace(args.filePath)
		}

		if !strings.HasSuffix(pathAsNameSpace, "."+namespace) {
			return fileAndLineError(args.filePath, lineNum+1, "Namespace "+namespace+" is not a suffix of "+pathAsNameSpace)
		}
	}

	return ""
}

func getFilePathAsNameSpace(filePath string) string {
	filePath = strings.Replace(filePath, "\\", "/", -1)

	base := strings.LastIndex(filePath, "/")
	dir := filePath[0:base]

	colon := strings.IndexRune(dir, ':')
	if colon >= 0 {
		dir = dir[colon+1:]
	}

	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}

	pathAsNameSpace := strings.Replace(dir, "/", ".", -1)
	pathAsNameSpace = strings.Replace(pathAsNameSpace, "\\", ".", -1)
	return pathAsNameSpace
}

func ruleCheckBadClassName(args ruleCheckArgs) string {
	base := filepath.Base(args.filePath)
	base = base[0 : len(base)-len(".cs")]

	for lineNum, line := range args.fileAsLines {
		line = strings.Trim(line, " \t")

		className := ""
		if strings.HasPrefix(line, "class ") {
			className = line[len("class "):]
		} else if strings.HasPrefix(line, "public class ") {
			className = line[len("public class "):]
		} else if strings.HasPrefix(line, "internal class ") {
			className = line[len("internal class "):]
		} else {
			continue
		}

		if className != base {
			return fileAndLineError(args.filePath, lineNum+1, "Class name "+className+" should be "+base+" instead")
		}
	}

	return ""
}

func ruleCheckNoMultiplePublicClasses(args ruleCheckArgs) string {
	count := 0
	for lineNum, line := range args.fileAsLines {
		line = strings.Trim(line, " \t")

		if strings.HasPrefix(line, "public class ") {
			count++
			if count > 1 {
				return fileAndLineError(args.filePath, lineNum+1, "Cannot have multiple public classes per file")
			}
		}
	}

	return ""
}

func ruleCheckWindowsNewlines(args ruleCheckArgs) string {
	s := strings.Replace(args.fileAsString, "\r\n", "", -1)
	if strings.ContainsRune(s, '\n') {
		return fileError(args.filePath, "File contains non-Windows (Linux) newlines")
	}
	if strings.ContainsRune(s, '\r') {
		return fileError(args.filePath, "File contains non-Windows (Old Mac) newlines")
	}
	return ""
}

func ruleCheckLinuxNewlines(args ruleCheckArgs) string {
	if strings.Contains(args.fileAsString, "\r\n") {
		return fileError(args.filePath, "File contains non-Linux (Windows) newlines")
	}
	s := strings.Replace(args.fileAsString, "\r\n", "", -1)
	if strings.ContainsRune(s, '\r') {
		return fileError(args.filePath, "File contains non-Linux (Old Mac) newlines")
	}
	return ""
}

func ruleCheckOldMacNewlines(args ruleCheckArgs) string {
	if strings.Contains(args.fileAsString, "\r\n") {
		return fileError(args.filePath, "File contains non-Old Mac (Windows) newlines")
	}
	s := strings.Replace(args.fileAsString, "\r\n", "", -1)
	if strings.ContainsRune(s, '\n') {
		return fileError(args.filePath, "File contains non-Old Mac (Linux) newlines")
	}
	return ""
}

func ruleCheckNeedSpaceAfterKeyword(args ruleCheckArgs) string {
	var buffer bytes.Buffer

	for lineNum, line := range args.fileAsLines {
		line = strings.Trim(line, " \t")

		for _, keyword := range csharpKeywordsWithSpacedParens {
			if strings.HasPrefix(line, keyword+"(") {
				buffer.WriteString(fileAndLineError(args.filePath, lineNum+1, "Need space between keyword "+keyword+" and open paren\n"))
				continue
			}
		}
	}

	return buffer.String()
}

/**************************************************************************/

// Get list of files.

func getListOfFiles() []string {
	if len(os.Args) > 1 {
		return getListOfFilesFromArguments()
	}
	return getListOfFilesFromGit()
}

func getListOfFilesFromArguments() []string {
	// Add all arguments including handling for globs.
	var files []string

	for i := 1; i < len(os.Args); i++ {
		pattern := os.Args[i]
		if strings.ContainsAny(pattern, "?*") {
			matches, err := filepath.Glob(pattern)
			if err == nil {
				files = append(files, matches...)
			}
		} else {
			files = append(files, pattern)
		}
	}

	// Add all files under all subdirs.
	var files2 []string

	for len(files) > 0 {
		f := files[len(files)-1]
		files = files[0 : len(files)-1]

		fileInfo, err := os.Stat(f)
		if err != nil {
			continue
		}

		if fileInfo.Mode().IsRegular() {
			files2 = append(files2, f)
		} else {
			if filepath.Base(f) == ".git" {
				continue
			}

			filePtr, err := os.Open(f)
			if err != nil {
				continue
			}

			fileNames, err := filePtr.Readdirnames(-1)
			sort.Strings(fileNames)
			for _, fn := range fileNames {
				fn = path.Join(f, fn)
				files = append(files, fn)
			}
		}
	}

	stringArrayReverse(files2)

	return files2
}

func getListOfFilesFromGit() []string {
	gotoGitRepoRootDir()

	// Get list of modified and staged files from git.
	diff1 := execAndGetOutput("git", []string{"diff", "--name-only", "--cached"})

	// Get list of modified but unstaged files from git.
	diff2 := execAndGetOutput("git", []string{"diff", "--name-only"})

	// Get list of new files from git.
	diff3 := execAndGetOutput("git", []string{"ls-files", "--others", "--exclude-standard"})

	array1 := convertStringToLines(diff1, true)
	array2 := convertStringToLines(diff2, true)
	array3 := convertStringToLines(diff3, true)

	return append(append(array1, array2...), array3...)
}

func gotoGitRepoRootDir() {
	filePath, _ := filepath.Abs(".")
	filePath = strings.Replace(filePath, "\\", "/", -1)
	paths := strings.Split(filePath, "/")

	for len(paths) > 0 {
		d := strings.Join(paths, "/")
		p := d + "/.git"

		fileInfo, err := os.Stat(p)
		if err == nil && fileInfo.IsDir() {
			os.Chdir(d)
			return
		}

		paths = paths[0 : len(paths)-1]
	}

	fatal("Not a git repository")
}

/**************************************************************************/

// Run rules.

func runRulesOnFiles(rules []rule, files []string) {
	for _, filePath := range files {
		fileExtension := filepath.Ext(filePath)
		filePath := getCanonicalPath(filePath)
		fileLoaded := false
		fileAsString := ""
		var fileAsLines []string

		firstNChars := readFirstNChars(filePath, 50)
		isBinary := hasControlCharacters(firstNChars)

		for _, rule := range rules {
			// Skip by file extension.
			if len(rule.fileExtensions) != 0 && !stringArrayContains(rule.fileExtensions, fileExtension) {
				continue
			}

			// Skip binary files.
			if isBinary && (rule.fileTypeFlags&flagBinaryFile == 0) {
				continue
			}

			// Skip text files.
			if !isBinary && (rule.fileTypeFlags&flagTextFile == 0) {
				continue
			}

			// Lazy load files.
			if !fileLoaded {
				fileAsString = readFileString(filePath)
				fileAsLines = convertStringToLines(fileAsString, false)
				fileLoaded = true
			}

			// Perform rule check.
			msg := rule.checkFunc(ruleCheckArgs{
				filePath:     filePath,
				fileAsString: fileAsString,
				fileAsLines:  fileAsLines,
			})

			// Print error message if any.
			if msg != "" {
				fmt.Println(msg)
			}
		}
	}
}

/**************************************************************************/

// Main.

func main() {
	rules := loadConfigFile()
	files := getListOfFiles()

	if len(files) == 0 {
		return
	}

	runRulesOnFiles(rules, files)
}

/**************************************************************************/

// End.
