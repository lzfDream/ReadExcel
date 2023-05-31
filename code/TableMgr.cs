// 生成
using Table;
using System.Collections.Generic;

interface ITableModule {
    void load();
}

public class TableMgr {
    private List<ITableModule> tables;
    private static TableMgr instance;

    private TableMgr() {
        tables.Add(new item())
        tables.Add(new test())
    }

    public static TableMgr Instance() {
        if (instance == null) {
            instance = new TableMgr();
        }
        return instance;
    }

    public void load() {
        foreach(ITableModule table in tables) {
            table.load();
        }
    }
}
