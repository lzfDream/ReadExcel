// 由github.com/lzfDream/ReadExcel生成, 请勿修改
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;

namespace Table
{
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

        public Dictionary<int, Item> AllItem;

        public void Load(in string path)
        {
            string json = File.ReadAllText(path + "/item.json");
            AllItem = JsonSerializer.Deserialize<Dictionary<int, Item>>(json);
        }

        public Item Get(int Id)
        {
            var dict0 = AllItem;
            if (!dict0.TryGetValue(Id, out var dict1))
            {
                Debug.TableErrorLog(string.Format("get TableItem data fail, key Id: {0}", Id));
                return default;
            }
            return dict1;
        }

        public Dictionary<int, Item> GatAllItem()
        {
            return AllItem;
        }
    }
}
