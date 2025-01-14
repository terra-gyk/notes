takeown /F "C:\Recovery\*" /R /A
cacls "C:\Recovery\*.*" /T /grant admin:F
rmdir /S /Q "C:\Recovery\"