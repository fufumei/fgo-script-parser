if (Test-Path  ".\script-length.md") {
	Remove-Item  ".\script-length.md"
}

echo "|Chapter|Lines|`n|---|---|" | Out-file -FilePath ".\script-length.md"
foreach ($Folder in Get-ChildItem $args[0]) {
	$folderName=$Folder.BaseName
	$matches=0
	foreach ($File in Get-ChildItem -Path $Folder -File -Recurse) {
		$matches += .\fgo-script-parser $File
	}
	echo "|${folderName}|${matches}|" | Out-file -FilePath ".\script-length.md" -Append
}