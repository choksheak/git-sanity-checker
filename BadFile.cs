# DoNothing
This check will always pass.

# NoTabs
    There is one tab before this line.

# NoLeadingSpaces
 There is one space before this line.

# TabsVsSpacesOnly
The above two lines will produce this error.

# ConsistentNewlines
This line ends with CR.This line ends with LF.
This line ends with CR/LF.

# ConsistentIndentWidth
  Two space indent.
   Three space indent.
    Four space indent.
     Five space indent.

# BadNameSpace
namespace bogus.ns

# BadClassName
public class BogusPublicClass
internal class BogusInternalClass

# NoMultiplePublicClasses
public class PublicClass1
public class PublicClass2

# WindowsNewlines
Scenario already covered under ConsistentNewlines.

# LinuxNewlines
Scenario already covered under ConsistentNewlines.

# OldMacNewlines
Scenario already covered under ConsistentNewlines.

# NeedSpaceAfterKeyword
catch()
for()
foreach()
if()
lock()
switch()
using()
while()
catch ()
for ()
foreach ()
if ()
lock ()
switch ()
using ()
while ()
