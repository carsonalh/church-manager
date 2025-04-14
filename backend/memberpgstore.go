package main

// An in-memory store for now
type MemberPgStore struct {
	members []Member
}

func CreateMemberPgStore() *MemberPgStore {
	return &MemberPgStore{members: make([]Member, 0)}
}

func (store *MemberPgStore) Save(member *Member) error {
	if member.Id != nil {
		for i, m := range store.members {
			if *m.Id == *member.Id {
				store.members[i] = *member
				break
			}
		}
	} else {
		var nextId *uint64
		maxId := uint64(0)

		for _, m := range store.members {
			maxId = max(maxId, *m.Id)
		}

		nextId = new(uint64)
		*nextId = maxId + 1
		member.Id = nextId
		store.members = append(store.members, *member)
	}

	return nil
}

func (store *MemberPgStore) GetById(id uint64) (*Member, error) {
	for _, m := range store.members {
		if *m.Id == id {
			ret := new(Member)
			*ret = m
			return ret, nil
		}
	}

	return nil, nil
}

func (store *MemberPgStore) Get(pageSize uint, page uint) ([]Member, error) {
	if page*pageSize > uint(len(store.members)) {
		return []Member{}, nil
	}

	start := page * pageSize
	end := (page + 1) * pageSize

	end = min(end, uint(len(store.members)))

	return store.members[start:end], nil
}

func (store *MemberPgStore) Delete(id uint64) (bool, error) {
	for i, m := range store.members {
		if *m.Id == id {
			store.members = append(store.members[:i], store.members[i+1:]...)
			return true, nil
		}
	}

	return false, nil
}
