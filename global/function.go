package global

func (p *PMapRole) ID() int64 {
	return p.RoleID
}

func (p *PMapRole) Type() int8 {
	return ACTOR_ROLE
}

func (p *PMapRole) GetPos() *PPos {
	return p.Pos
}
func (p *PMapRole) GetMove() (int32, *PMovePath, *PPos) {
	return p.MoveSpeed, p.MovePath, p.Pos
}

func (p *PMapRole) SetMovePath(path *PMovePath) {
	p.MovePath = path
}

func (p *PMapRole) IsAlive() bool {
	return p.State != 1
}
func (p *PMapRole) GetRadius() float64 {
	return 0.5
}

func (p *PMapRole)GetCamp() int8 {
	return p.Camp
}

func (p *PMapRole)GetBuffs() []*PBuff {
	return p.Buffs
}

func (p *PMapRole)SetBuffs(buff []*PBuff){
	p.Buffs = buff
}

//-------------------------------------monster---------------------------------
func (p *PMapMonster) ID() int64 {
	return p.MonsterID
}

func (p *PMapMonster) Type() int8 {
	return ACTOR_MONSTER
}

func (p *PMapMonster) GetPos() *PPos {
	return p.Pos
}

func (p *PMapMonster) GetMove() (int32, *PMovePath, *PPos) {
	return p.MoveSpeed, p.MovePath, p.Pos
}
func (p *PMapMonster) SetMovePath(path *PMovePath) {
	p.MovePath = path
}
func (p *PMapMonster) IsAlive() bool {
	return p.HP > 0
}

func (p *PMapMonster) GetRadius() float64 {
	return float64(p.Radius)
}

func (p *PMapMonster)GetCamp() int8 {
	return p.Camp
}

func (p *PMapMonster)GetBuffs() []*PBuff {
	return p.Buffs
}

func (p *PMapMonster)SetBuffs(buff []*PBuff){
	p.Buffs = buff
}
//--------------------------------------------------------------------------------
