package helper

//ID 生成器接口
type IdGeneratorIntfs interface {
    GetUint64() uint64 //获得一个uint64类型的ID
}