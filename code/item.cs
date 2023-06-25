// 生成
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;

public class TableItem : ITableModule
{
    public class Item
    {
        public int Id { get; set; }
        public string Name { get; set; }
        public string Desc { get; set; }
        public int Price { get; set; }
        public int UpgradeToItemId { get; set; }
        public bool BatchUseable { get; set; }
        public string ExchangeStream { get; set; }
        public string ExchangeList { get; set; }
        public int ExchangeColumn { get; set; }
    }

    public Dictionary<string, Item> AllItem;

    public void Load(in string path)
    {
        string json = File.ReadAllText(path + "/item.json");
        AllItem = JsonSerializer.Deserialize<Dictionary<string, Item>>(json);
    }
}
