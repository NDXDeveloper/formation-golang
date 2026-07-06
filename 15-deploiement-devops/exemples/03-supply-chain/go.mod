module github.com/exemple/supplychain

go 1.25.0

// ⚠️ Dépendance VOLONTAIREMENT vulnérable (golang.org/x/text@v0.3.7 porte
//    GO-2022-1059) pour démontrer govulncheck. Ne jamais faire en production.
require golang.org/x/text v0.3.7
