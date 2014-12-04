package types

type Fleet struct {
	table map[string]Ship
}

func (fleet *Fleet) Add(value Ship) {
	if fleet.Find(value.Id).Id == "" {
		fleet.table[value.Id] = value
	} else {
		fleet.Update(value)
	}
}

func (fleet *Fleet) Find(idOrName string) Ship {
	for _, ship := range fleet.table {
		if ship.Name == idOrName || ship.Fqdn == idOrName || ship.Id == idOrName {
			return ship
		}
	}
	return Ship{}
}

func (fleet *Fleet) Update(ship Ship) {
	fleet.table[ship.Id] = ship
}

func (fleet *Fleet) Delete(ship Ship) {
	delete(fleet.table, ship.Id)
}

func (fleet *Fleet) Append(ships []Ship) {
	if fleet.table == nil {
		fleet.table = make(map[string]Ship)
	}
	if ships != nil {
		for _, ship := range ships {
			fleet.table[ship.Id] = ship
		}
	}
}

func (fleet *Fleet) AppendShip(ship Ship) {
	fleet.table[ship.Id] = ship
}

func (fleet *Fleet) Available() []Ship {

	v := make([]Ship, 0)

	for _, value := range fleet.table {
		if value.LocalPort > 0 {
			v = append(v, value)
		}
	}
	return v
}

func (fleet *Fleet) Ships() []Ship {
	v := make([]Ship, 0, len(fleet.table))

	for _, value := range fleet.table {
		v = append(v, value)
	}
	return v
}

func NewFleet(ships []Ship) Fleet {
	var fleet Fleet
	fleet.Append(ships)
	return fleet
}
