package cuckoo

import (
	"bytes"
)

func mod(code uint32, bucket_size int) uint32 {
	//return code % uint(bucket_size)
	return code & (uint32(bucket_size) - 1) //bucket_size == 2^n
}
func (tb *hashTable) findAndKill(key []byte, kill bool) bool {
	for i := 0; i < WAYS; i++ {
		var idx = (tb.idx + i) % WAYS
		var table = &tb.core[idx]
		var code = table.hash(key)
		var index = mod(code, len(table.bucket))
		var target = table.bucket[index]
		if target != nil &&
			target.code[idx] == code &&
			bytes.Compare(target.key, key) == 0 {
			if kill {
				table.bucket[index] = nil
			}
			return true
		}
	}
	return false
}

func (tb *hashTable) Search(key []byte) bool {
	return tb.findAndKill(key, false)
}
func (tb *hashTable) Remove(key []byte) bool {
	var killed = tb.findAndKill(key, true)
	if killed {
		tb.cnt--
	}
	return killed
}

func (tb *hashTable) Insert(key []byte) bool {
	if !tb.findAndKill(key, false) {
		tb.cnt++

		var unit = new(node)
		unit.key = key
		for i := 0; i < WAYS; i++ {
			unit.code[i] = tb.core[i].hash(key)
		}

		for obj, age := unit, 0; ; age++ {
			for idx, trys := tb.idx, 0; trys < WAYS; idx = (idx + 1) % WAYS {
				var table = &tb.core[idx]
				var index = mod(obj.code[idx], len(table.bucket))
				if table.bucket[index] == nil {
					table.bucket[index] = obj
					return true
				}
				obj, table.bucket[index] = table.bucket[index], obj
				if obj == unit {
					trys++ //回绕计数
				}
			}

			if age != 0 { //这里设定一个阈值，限制扩容次数
				panic("too many conflicts")
			} //实际上不能解决大量hash重码的情况，最坏情况只能报错

			tb.expand() //调整失败(回绕)，扩容
		}
	}
	return false
}
func (tb *hashTable) expand() {
	tb.idx = (tb.idx + (WAYS - 1)) % WAYS
	var table = &tb.core[tb.idx]
	var old_bucket = table.bucket
	table.bucket = make([]*node, len(old_bucket)<<WAYS)
	for _, u := range old_bucket {
		if u != nil {
			var index = mod(u.code[tb.idx], len(table.bucket))
			table.bucket[index] = u //倍扩，绝对不会冲突
		}
	}
}
