package tree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

type NodeValue interface {
	IsEqual(other NodeValue) bool
	GetKey() interface{}
	GetKeyCode() []byte

	/*Compare the result will be 0 if this == other, -1 if this < other, and +1 if this > other*/
	Compare(other NodeValue) int
}

type NodeValueCommS struct {
	V string
}

type NodeValueCommI struct {
	V int
}

type Node struct {
	Value    NodeValue
	Parent   *Node
	Children []*Node
}

// LinkParent 关联父节点
func LinkParent(root *Node) {
	PostorderTraversal(root, func(current *Node) bool {
		if current.Children != nil {
			for _, child := range current.Children {
				child.Parent = current
			}
		}
		return false
	})
}

// PreorderTraversal 前序遍历函数
func PreorderTraversal(root *Node, handle func(current *Node) bool /*broken*/) bool /*broken*/ {
	if root == nil {
		return false
	}

	if handle(root) {
		return true
	}

	if root.Children != nil {
		for _, child := range root.Children {
			if PreorderTraversal(child, handle) {
				return true
			}
		}
	}

	return false
}

// PostorderTraversal 后序遍历函数
func PostorderTraversal(root *Node, handle func(current *Node) bool /*broken*/) bool /*broken*/ {
	if root == nil {
		return false
	}

	if root.Children != nil {
		for _, child := range root.Children {
			if PostorderTraversal(child, handle) {
				return true
			}
		}
	}

	return handle(root)
}

// PreorderFind 前序搜索函数: 从root中找到target
func PreorderFind(root *Node, targetValue NodeValue) *Node {
	if root == nil {
		return nil
	}

	if root.Value.IsEqual(targetValue) {
		return root
	}

	if root.Children != nil {
		for _, child := range root.Children {
			target := PreorderFind(child, targetValue)
			if target != nil {
				return target
			}
		}
	}

	return nil
}

// PostorderFind 后序搜索函数: 从root中找到target
func PostorderFind(root *Node, targetValue NodeValue) *Node {
	if root == nil {
		return nil
	}

	if root.Children != nil {
		for _, child := range root.Children {
			target := PostorderFind(child, targetValue)
			if target != nil {
				return target
			}
		}
	}

	if root.Value.IsEqual(targetValue) {
		return root
	}

	return nil
}

// PreorderFullPath 前序遍历所有根到末端的路径
func PreorderFullPath(root *Node) [][]*Node {
	var result [][]*Node
	preorderFullPathRec(root, []*Node{}, &result)
	return result
}

func preorderFullPathRec(root *Node, path []*Node, result *[][]*Node) {
	if root == nil {
		return
	}

	path = append(path, root)

	if len(root.Children) == 0 {
		*result = append(*result, append([]*Node{}, path...))
		return
	}

	for _, child := range root.Children {
		preorderFullPathRec(child, path, result)
	}
}

// SortBrother 兄弟节点排序
func SortBrother(root *Node, descOrder bool) {
	if root == nil || root.Children == nil {
		return
	}

	sort.Slice(root.Children, func(i, j int) bool {
		smaller := root.Children[i].Value.Compare(root.Children[j].Value) < 0
		if descOrder {
			return !smaller
		} else {
			return smaller
		}
	})

	for _, child := range root.Children {
		SortBrother(child, descOrder)
	}
}

// MergeIntersectSubTrees 合并两个相交的子树
func MergeIntersectSubTrees(tree1, tree2 *Node) *Node {
	if tree1 == nil {
		return tree2
	}
	if tree2 == nil {
		return tree1
	}

	node2 := PreorderFind(tree1, tree2.Value)
	if node2 != nil {
		MergeChildren(tree2, node2)
		return tree1
	}

	node1 := PreorderFind(tree2, tree1.Value)
	if node1 != nil {
		MergeChildren(tree1, node1)
		return tree2
	}

	return nil
}

// MergeChildren 合并两个节点的子节点
func MergeChildren(src, dst *Node) {
	PreorderTraversal(src, func(current *Node) bool {
		if PreorderFind(dst, current.Value) == nil {
			if current.Parent == nil {
				panic("logical error, dst has the same root with src")
			}
			parent := PreorderFind(dst, current.Parent.Value)
			if parent == nil {
				panic(fmt.Sprintf("logical error, dst has no the parent#%v", current.Parent.Value))
			}

			added := &Node{Value: current.Value, Parent: parent, Children: current.Children}
			parent.Children = append(parent.Children, added)
		}
		return false
	})
}

// MergeChildren2 合并两个节点的子节点(排重)(错误版)
func MergeChildren2(src, dst *Node) {
	if src.Children != nil {
		for _, srcChild := range src.Children {
			has := false
			if dst.Children != nil {
				for _, dstChild := range dst.Children {
					if srcChild.Value.IsEqual(dstChild.Value) {
						has = true
						break
					}
				}
			}
			if !has {
				dst.Children = append(dst.Children, srcChild)
			}
		}
	}
}

// Contains 判断是否包含该子树(递归遍历树)
func (n *Node) Contains(sub *Node) bool {
	if n == nil {
		return false
	} else if sub == nil {
		return true
	}

	// 查找相同的起始结点
	var first *Node
	PreorderTraversal(n, func(current *Node) bool {
		if current.Value.IsEqual(sub.Value) {
			first = current
			return true
		}
		return false
	})

	if first == nil {
		return false
	}

	return first.containsRecursive(sub)
}

// Contains2 判断是否包含该子树(且兄弟顺序一致)(比较树结点队列)
func (n *Node) Contains2(sub *Node) bool {
	if n == nil {
		return false
	} else if sub == nil {
		return true
	}

	// 遍历两个树，生成结点列表
	var values1 []NodeValue
	var values2 []NodeValue
	PreorderTraversal(n, func(current *Node) bool {
		values1 = append(values1, current.Value)
		return false
	})
	PreorderTraversal(sub, func(current *Node) bool {
		values2 = append(values2, current.Value)
		return false
	})
	if len(values1) < len(values2) {
		return false
	}

	// 判断values1是否包含values2
	i2 := 0
	hasFist := false
	for _, value := range values1 {
		if hasFist {
			if value.IsEqual(values2[i2]) {
				i2 += 1
			} else {
				break
			}
		} else {
			if value.IsEqual(values2[i2]) {
				hasFist = true
				i2 += 1
			}
		}

		if i2 >= len(values2) {
			break
		}
	}

	return hasFist && i2 >= len(values2)
}

// Contains3 判断是否包含该子树(!!!有bug)
func (n *Node) Contains3(sub *Node) bool {
	if n == nil {
		return false
	} else if sub == nil {
		return true
	}

	nPaths := PreorderFullPath(n)
	sPaths := PreorderFullPath(sub)
	if len(nPaths) < len(sPaths) {
		return false
	}

	sortPathsByValue(nPaths)
	sortPathsByValue(sPaths)

	for _, sPath := range sPaths {
		contains := false
		for _, nPath := range nPaths {
			if len(nPath) < len(sPath) {
				continue
			}

			eql := true
			for idx, sNode := range sPath {
				nNode := nPath[idx]
				if !nNode.Value.IsEqual(sNode.Value) {
					eql = false
					break
				}
			}
			if eql {
				contains = true
				break
			}
		}

		if !contains {
			return false
		}
	}

	return true
}

// Contains4 判断是否包含该子树
func (n *Node) Contains4(sub *Node) bool {
	if n == nil {
		return false
	} else if sub == nil {
		return true
	}

	nPaths := PreorderFullPath(n)
	sPaths := PreorderFullPath(sub)
	if len(nPaths) < len(sPaths) {
		return false
	}

	sortPathsByValue(nPaths)
	sortPathsByValue(sPaths)

	nBytesPaths := transBytesPaths(nPaths)
	sBytesPaths := transBytesPaths(sPaths)

	for _, sPath := range sBytesPaths {
		contains := false

		for _, nPath := range nBytesPaths {
			contains = bytes.Contains(nPath, sPath)
			if contains {
				break
			}
		}

		if !contains {
			return false
		}
	}

	return true
}

// DeleteNode 从tree中删除node
//
//	其子节点挂载到其父节点上
//	若不存在父节点,则其他子节点挂载到兄节点
func DeleteNode(tree, node *Node) *Node {
	if node == nil {
		return tree
	}
	if tree == nil {
		return nil
	}
	if tree.Parent != nil {
		panic("tree is not root node")
	}

	if tree == node {
		if tree.Children == nil {
			return nil
		}
		sub := len(tree.Children)
		if sub == 0 {
			return nil
		}

		older := tree.Children[0]
		youngers := tree.Children[1:]
		older.Children = append(older.Children, youngers...)
		for _, younger := range youngers {
			younger.Parent = older
		}

		node.Children = nil
		return older
	} else {
		for i, child := range node.Parent.Children {
			if child == node {
				if i == 0 {
					node.Parent.Children = node.Parent.Children[1:]
				} else if i == (len(node.Parent.Children) - 1) {
					node.Parent.Children = node.Parent.Children[:i]
				} else {
					node.Parent.Children = append(node.Parent.Children[:i], node.Parent.Children[i+1:]...)
				}
				break
			}
		}

		if node.Children != nil {
			node.Parent.Children = append(node.Parent.Children, node.Children...)
			for _, child := range node.Children {
				child.Parent = node.Parent
			}
		}

		node.Parent = nil
		node.Children = nil
		return tree
	}
}

func sortPathsByValue(paths [][]*Node) {
	sort.Slice(paths, func(i, j int) bool {
		pathi := paths[i]
		pathj := paths[j]
		cnt := len(pathi)
		if len(pathi) > len(pathj) {
			cnt = len(pathj)
		}

		for idx := 0; idx < cnt; idx += 1 {
			nodei := pathi[idx]
			nodej := pathj[idx]
			cmp := nodei.Value.Compare(nodej.Value)
			if cmp != 0 {
				return cmp < 0
			}
		}
		return true
	})
}

func transBytesPaths(paths [][]*Node) [][]byte {
	if paths == nil {
		return nil
	}

	result := make([][]byte, 0, len(paths))
	for _, path := range paths {
		var builder bytes.Buffer
		for _, node := range path {
			builder.Write(node.Value.GetKeyCode())
			builder.WriteByte(0xff)
		}
		result = append(result, builder.Bytes())
	}

	return result
}

func (n *Node) containsRecursive(sub *Node) bool {
	if n == nil || sub == nil {
		// 如果当前节点为 nil，表示子树 sub 在此处已经完全匹配，返回 true
		return true
	} else if !n.Value.IsEqual(sub.Value) {
		// 如果当前节点的值与 sub 的值不相等，则返回 false
		return false
	} else if sub.Children == nil {
		// 子树已经结束
		return true
	} else if n.Children == nil {
		return len(sub.Children) == 0
	}

	sl := len(sub.Children)
	nl := len(n.Children)
	if sl > nl {
		return false
	}

	// 遍历当前节点的子节点，递归地判断子节点及其子树是否包含子树 sub
	for i := range sub.Children {
		if !n.Children[i].containsRecursive(sub.Children[i]) {
			return false
		}
	}

	// 如果所有子节点都包含子树 sub 的对应子节点，则返回 true
	return true
}

// ListNode 带父节点信息的列表节点
type ListNode struct {
	Value     NodeValue
	ParentKey interface{}
}

// BuildTreeFromList 从节点列表生成节点树
func BuildTreeFromList(listNodes []*ListNode) *Node {
	nodeMap := make(map[interface{}]*Node)

	// 创建树节点并建立节点映射
	for _, listNode := range listNodes {
		treeNode := &Node{
			Value:    listNode.Value,
			Parent:   nil,
			Children: nil,
		}
		nodeMap[listNode.Value.GetKey()] = treeNode
	}

	// 建立父子关系
	for _, listNode := range listNodes {
		treeNode := nodeMap[listNode.Value.GetKey()]
		if parentNode, exists := nodeMap[listNode.ParentKey]; exists {
			treeNode.Parent = parentNode
			parentNode.Children = append(parentNode.Children, treeNode)
		}
	}

	// 查找并返回根节点
	var root *Node
	for _, treeNode := range nodeMap {
		if treeNode.Parent == nil {
			root = treeNode
			break
		}
	}

	return root
}

func (cv *NodeValueCommS) IsEqual(other NodeValue) bool {
	if otherS, ok := other.(*NodeValueCommS); ok {
		return cv.V == otherS.V
	}
	return false
}

func (cv *NodeValueCommS) GetKey() interface{} {
	return cv.V
}

func (cv *NodeValueCommS) GetKeyCode() []byte {
	return []byte(cv.V)
}

func (cv *NodeValueCommS) Compare(other NodeValue) int {
	o, ok := other.(*NodeValueCommS)
	if !ok {
		return -2
	}
	return strings.Compare(cv.V, o.V)
}

func (cv *NodeValueCommI) IsEqual(other NodeValue) bool {
	if otherI, ok := other.(*NodeValueCommI); ok {
		return cv.V == otherI.V
	}
	return false
}

func (cv *NodeValueCommI) GetKey() interface{} {
	return cv.V
}

func (cv *NodeValueCommI) GetKeyCode() []byte {
	code := make([]byte, 4)
	binary.BigEndian.PutUint32(code, uint32(cv.V))
	return code
}

func (cv *NodeValueCommI) Compare(other NodeValue) int {
	o, ok := other.(*NodeValueCommI)
	if !ok {
		return -2
	} else if cv.V == o.V {
		return 0
	} else if cv.V > o.V {
		return 1
	} else {
		return -1
	}
}
