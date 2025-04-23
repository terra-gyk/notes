takeown /F "C:\Windows.old\*" /R /A
cacls "C:\Windows.old\*.*" /T /grant admin:F
rmdir /S /Q "C:\Windows.old\"