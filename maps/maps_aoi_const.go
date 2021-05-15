package maps

type le struct {
	leave []Area
	enter []Area
}

// 预先构造一个数据，方便计算出结果,相比取补集来说，这样快2倍以上
var constLeaveEnterMap = map[Area]*le{
	{1,1}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{1,-1}},
		enter:[]Area{{-1,1},{0,1},{1,-1},{1,0},{1,1}}},
	{-2,-2}:{leave:[]Area{{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,-1},{1,0}}},
	{-1,-1}:{leave:[]Area{{-1,1},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{1,-1}}},
	{-1,1}:{leave:[]Area{{-1,-1},{0,-1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,1},{1,1}}},
	{0,2}:{leave:[]Area{{-1,-1},{-1,0},{0,-1},{0,0},{1,-1},{1,0}},
		enter:[]Area{{-1,0},{-1,1},{0,0},{0,1},{1,0},{1,1}}},
	{2,-1}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,1}},
		enter:[]Area{{-1,-1},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}}},
	{2,0}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1}},
		enter:[]Area{{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}}},
	{-2,2}:{leave:[]Area{{-1,-1},{-1,0},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,0},{1,1}}},
	{-1,-2}:{leave:[]Area{{-1,0},{-1,1},{0,0},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{1,-1},{1,0}}},
	{-1,2}:{leave:[]Area{{-1,-1},{-1,0},{0,-1},{0,0},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,0},{0,1},{1,0},{1,1}}},
	{0,1}:{leave:[]Area{{-1,-1},{0,-1},{1,-1}},
		enter:[]Area{{-1,1},{0,1},{1,1}}},
	{1,2}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{1,-1},{1,0}},
		enter:[]Area{{-1,0},{-1,1},{0,0},{0,1},{1,-1},{1,0},{1,1}}},
	{-2,-1}:{leave:[]Area{{-1,1},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,-1}}},
	{-2,1}:{leave:[]Area{{-1,-1},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,1}}},
	{-1,0}:{leave:[]Area{{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1}}},
	{1,-1}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,1},{1,1}},
		enter:[]Area{{-1,-1},{0,-1},{1,-1},{1,0},{1,1}}},
	{1,0}:{leave:[]Area{{-1,-1},{-1,0},{-1,1}},
		enter:[]Area{{1,-1},{1,0},{1,1}}},
	{2,-2}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}}},
	{2,1}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,-1}},
		enter:[]Area{{-1,1},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}}},
	{-2,0}:{leave:[]Area{{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1}}},
	{0,-2}:{leave:[]Area{{-1,0},{-1,1},{0,0},{0,1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{0,-1},{0,0},{1,-1},{1,0}}},
	{0,-1}:{leave:[]Area{{-1,1},{0,1},{1,1}},
		enter:[]Area{{-1,-1},{0,-1},{1,-1}}},
	{1,-2}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,0},{0,1},{1,0},{1,1}},
		enter:[]Area{{-1,-1},{-1,0},{0,-1},{0,0},{1,-1},{1,0},{1,1}}},
	{2,2}:{leave:[]Area{{-1,-1},{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,-1},{1,0}},
		enter:[]Area{{-1,0},{-1,1},{0,-1},{0,0},{0,1},{1,-1},{1,0},{1,1}}},

}
