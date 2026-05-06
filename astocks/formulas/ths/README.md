# 同花顺选股公式

## 文件说明

- `a_share_screener_basic.txt`: 基础版，可直接粘贴到同花顺条件选股公式中使用
- `a_share_screener_l2_placeholder.txt`: 含 `Level-2` 占位字段的完整版，替换占位表达式后使用

## 原始筛选要求

1. 近 15 天内有倍量突破
2. `1 > 大单净量 > 0`
3. `散户数量 < 0`
4. 多头排列
5. 至少 2 天突破前一天高点

说明：

- 基础版实现第 `1/4/5` 条
- 完整版在基础版之上增加第 `2/3` 条
- 第 `2/3` 条依赖同花顺 `Level-2` 指标

## 基础版条件

- 近 15 天内至少出现 1 次倍量突破
- 多头排列：`MA5 > MA10 > MA20 > MA60`
- 近 15 天内至少 2 天收盘突破前一天高点

## 完整版额外条件

- `0 < 大单净量 < 1`
- `散户数量 < 0`

## 使用方式

1. 在同花顺中打开“公式管理器”或“指标公式编辑器”
2. 新建“条件选股公式”
3. 复制对应文件内容并保存
4. 基础版可直接运行
5. 完整版先将以下两行替换为你本机同花顺 `Level-2` 可用字段

```text
BIG_ORDER_NET:=L2_BIGORDER_NET;
RETAIL_LEVEL:=L2_RETAIL;
```

## 如何查找 Level-2 字段名

1. 打开同花顺“公式管理器”
2. 新建一个测试用的指标公式，而不是直接改选股公式
3. 在系统函数、扩展行情、资金流向或 `Level-2` 分类里查找与“大单净量”“散户数量”名称接近的字段
4. 每次只测试一个字段，先写成最小公式并保存，确认是否能通过编译
5. 编译通过后，再把该字段替换到完整版公式里的两行占位代码

可按下面方式逐个测试：

```text
TEST:大单净量;
```

或：

```text
TEST:散户数量;
```

如果同花顺不接受中文指标名，通常说明这里需要的是系统内置字段名或某个现成指标公式名，而不是显示名称。

确认可用后，只替换这两行：

```text
BIG_ORDER_NET:=这里替换成大单净量的真实表达式;
RETAIL_LEVEL:=这里替换成散户数量的真实表达式;
```

如果 `Level-2` 字段不能直接在条件选股公式中调用，可先单独建两个指标公式，确保它们能正常输出，再尝试在完整版中引用对应结果。

## 基础版公式

```text
BREAK1:=C>REF(H,1);
DOUBLEVOL_BREAK:=BREAK1 AND VOL>=2*REF(VOL,1);
HAS_DOUBLEVOL_15:=COUNT(DOUBLEVOL_BREAK,15)>=1;
BULL_ALIGN:=MA(C,5)>MA(C,10) AND MA(C,10)>MA(C,20) AND MA(C,20)>MA(C,60);
BREAK_COUNT_15:=COUNT(BREAK1,15)>=2;
XG:HAS_DOUBLEVOL_15 AND BULL_ALIGN AND BREAK_COUNT_15;
```

## 完整版公式

```text
BREAK1:=C>REF(H,1);
DOUBLEVOL_BREAK:=BREAK1 AND VOL>=2*REF(VOL,1);
HAS_DOUBLEVOL_15:=COUNT(DOUBLEVOL_BREAK,15)>=1;
BULL_ALIGN:=MA(C,5)>MA(C,10) AND MA(C,10)>MA(C,20) AND MA(C,20)>MA(C,60);
BREAK_COUNT_15:=COUNT(BREAK1,15)>=2;

BIG_ORDER_NET:=L2_BIGORDER_NET;
RETAIL_LEVEL:=L2_RETAIL;

BIG_NET_OK:=BIG_ORDER_NET>0 AND BIG_ORDER_NET<1;
RETAIL_OK:=RETAIL_LEVEL<0;
XG:HAS_DOUBLEVOL_15 AND BULL_ALIGN AND BREAK_COUNT_15 AND BIG_NET_OK AND RETAIL_OK;
```
