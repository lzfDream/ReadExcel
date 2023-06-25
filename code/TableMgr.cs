// 生成
using Table;
using System.Collections.Generic;

interface ITableModule
{
    void Load(in string path);
}

public class TableMgr
{
    private List<ITableModule> tables;
    private static TableMgr instance;

    private TableMgr()
    {
        tables = new List<ITableModule>();
        tables.Add(new TableItem());
        tables.Add(new TableTest());
    }

    public static TableMgr Instance()
    {
        if (instance == null)
        {
            instance = new TableMgr();
        }
        return instance;
    }

    public void Load(in string path)
    {
        foreach(ITableModule table in tables)
        {
            table.Load(path);
        }
    }

    public T GetTable<T>() where T : class
    {
        Type type = typeof(T);
        foreach(ITableModule table in tables)
        {
            Type type2 = table.GetType();
            if (type == type2)
            {
                return table as T;
            }
        }
        return default(T);
    }
}
