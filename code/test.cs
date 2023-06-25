// 生成
using System;
using System.IO;
using System.Text.Json;

public class TableTest : ITableModule
{
    public class Item
    {
        public int test1 { get; set; }
        public double test3 { get; set; }
        public bool test4 { get; set; }
    }

    public Item KeyItem;

    public void Load(in string path)
    {
        string json = File.ReadAllText(path + "/test.json");
        KeyItem = JsonSerializer.Deserialize<Item>(json);
    }
}
