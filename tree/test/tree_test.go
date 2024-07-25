package test

import (
	"encoding/binary"
	"fmt"
	"github.com/patstar123/meeting-base/tree"
	"testing"
)

type V struct {
	v int
}

func (v V) IsEqual(other tree.NodeValue) bool {
	return v.v == other.(V).v
}

func (v V) GetKey() interface{} {
	return v.v
}

func (v V) GetKeyCode() []byte {
	code := make([]byte, 4)
	binary.BigEndian.PutUint32(code, uint32(v.v))
	return code
}

func (v V) Compare(other tree.NodeValue) int {
	o, ok := other.(V)
	if !ok {
		return -2
	} else if v.v == o.v {
		return 0
	} else if v.v > o.v {
		return 1
	} else {
		return -1
	}
}

func TestTree4(t *testing.T) {
	// 构建一个普通树示例
	root := &tree.Node{
		Value: V{1},
		Children: []*tree.Node{
			{Value: V{2}, Children: []*tree.Node{
				{Value: V{4}, Children: nil},
				{Value: V{5}, Children: nil},
			}},
			{Value: V{3}, Children: []*tree.Node{
				{Value: V{6}, Children: nil},
				{Value: V{7}, Children: nil},
			}},
			{Value: V{8}, Children: nil},
		},
	}

	subTree := tree.GenLCASubTree(root, []tree.NodeValue{V{4}, V{6}, V{8}})
	fmt.Println("最小联通子树(4/6/8):")
	tree.PostorderTraversal(subTree, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	subTree = tree.GenLCASubTree(root, []tree.NodeValue{V{4}, V{5}, V{6}, V{8}})
	fmt.Println("最小联通子树(4/5/6/8):")
	tree.PostorderTraversal(subTree, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	subTree = tree.GenLCASubTree(root, []tree.NodeValue{V{4}, V{8}})
	fmt.Println("最小联通子树(4/8):")
	tree.PostorderTraversal(subTree, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	subTree = tree.GenLCASubTree(root, []tree.NodeValue{V{4}})
	fmt.Println("最小联通子树(4):")
	tree.PostorderTraversal(subTree, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	subTree = tree.GenLCASubTree(root, []tree.NodeValue{})
	fmt.Println("最小联通子树():")
	tree.PostorderTraversal(subTree, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()
}

func TestNodeList(t *testing.T) {
	listNodes := []*tree.ListNode{
		{V{1}, nil},
		{V{2}, 1},
		{V{3}, 1},
		{V{8}, 1},
		{V{4}, 2},
		{V{5}, 2},
		{V{6}, 3},
		{V{7}, 3},
	}

	root := tree.BuildTreeFromList(listNodes)

	fmt.Println("节点列表生成树:")
	tree.PostorderTraversal(root, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()
}

func TestContains(t *testing.T) {
	fmt.Println("节点列表生成树1:")
	list1 := []*tree.ListNode{
		{V{1}, nil},
		{V{2}, 1},
		{V{3}, 1},
	}
	tree1 := tree.BuildTreeFromList(list1)
	tree.PostorderTraversal(tree1, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	fmt.Println("节点列表生成树2:")
	list2 := []*tree.ListNode{
		{V{4}, nil},
		{V{5}, 4},
		{V{6}, 4},
	}
	tree2 := tree.BuildTreeFromList(list2)
	tree.PostorderTraversal(tree2, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	fmt.Println("节点列表生成树3:")
	list3 := []*tree.ListNode{
		{V{3}, nil},
		{V{4}, 3},
		{V{5}, 4},
		{V{6}, 4},
	}
	tree3 := tree.BuildTreeFromList(list3)
	tree.PostorderTraversal(tree3, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	fmt.Println("节点列表生成树4:")
	list4 := []*tree.ListNode{
		{V{4}, nil},
		{V{5}, 4},
		{V{6}, 4},
		{V{7}, 5},
	}
	tree4 := tree.BuildTreeFromList(list4)
	tree.PostorderTraversal(tree4, func(current *tree.Node) bool {
		fmt.Printf("%v ", current.Value.(V).v)
		return false
	})
	fmt.Println()

	fmt.Println("tree1 contains tree2?", tree1.Contains(tree2))
	fmt.Println("tree1 contains tree3?", tree1.Contains(tree3))
	fmt.Println("tree3 contains tree1?", tree3.Contains(tree1))
	fmt.Println("tree1 contains tree1?", tree1.Contains(tree1))
	fmt.Println("tree3 contains tree2?", tree3.Contains(tree2))
	fmt.Println("tree2 contains tree3?", tree2.Contains(tree3))
	fmt.Println("tree4 contains tree2?", tree3.Contains(tree2))
	fmt.Println("tree2 contains tree4?", tree2.Contains(tree3))
	fmt.Println("==========================")
	fmt.Println("tree1 contains tree2?", tree1.Contains2(tree2))
	fmt.Println("tree1 contains tree3?", tree1.Contains2(tree3))
	fmt.Println("tree3 contains tree1?", tree3.Contains2(tree1))
	fmt.Println("tree1 contains tree1?", tree1.Contains2(tree1))
	fmt.Println("tree3 contains tree2?", tree3.Contains2(tree2))
	fmt.Println("tree2 contains tree3?", tree2.Contains2(tree3))
	fmt.Println("tree4 contains tree2?", tree3.Contains2(tree2))
	fmt.Println("tree2 contains tree4?", tree2.Contains2(tree3))
}
