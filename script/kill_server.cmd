@echo ----------------------------- ---------------------------------------- ------------------------------------
@echo ----------------------------- --------------Start Kill All Server------------------------ ------------------
@echo ----------------------------- ---------------------------------------- ------------------------------------
TASKKILL /F /IM dungeon-match-server.exe
TASKKILL /F /IM game-server.exe
TASKKILL /F /IM gateway.exe
TASKKILL /F /IM gen-rank-server.exe
TASKKILL /F /IM guild-filter-server.exe
TASKKILL /F /IM new-battle-server.exe
TASKKILL /F /IM recommend-server.exe

@echo ----------------------------- ----------------Process Finished------------------------ -------------------
pause