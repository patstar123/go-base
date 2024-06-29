package tree

// FindLCA 查找树上两个节点的最小公共祖先(节点)
func FindLCA(root *Node, nodeValue1, nodeValue2 NodeValue) *Node {
	if root == nil || nodeValue1 == nil || nodeValue2 == nil {
		return nil
	}
	if root.Value.IsEqual(nodeValue1) || root.Value.IsEqual(nodeValue2) {
		return root
	}

	var lca *Node
	for _, child := range root.Children {
		childLCA := FindLCA(child, nodeValue1, nodeValue2)
		if childLCA != nil {
			if lca == nil {
				lca = childLCA
			} else {
				// 如果已经找到一个LCA，则表示当前节点是两个节点的公共祖先
				return root
			}
		}
	}

	return lca
}

// GenTwoNodesLCASubTree 生成树上两个节点的最小连通子树
func GenTwoNodesLCASubTree(root *Node, nodeValue1, nodeValue2 NodeValue) *Node {
	var node1, node2 *Node = nil, nil
	if nodeValue1 != nil {
		node1 = PostorderFind(root, nodeValue1)
	}
	if nodeValue2 != nil {
		node2 = PostorderFind(root, nodeValue2)
	}

	if node1 == nil && node2 == nil {
		return nil
	} else if node1 == nil {
		return &Node{
			Value:    nodeValue2,
			Parent:   nil,
			Children: nil,
		}
	} else if node2 == nil {
		return &Node{
			Value:    nodeValue1,
			Parent:   nil,
			Children: nil,
		}
	}

	lca := FindLCA(root, nodeValue1, nodeValue2)
	if lca == nil {
		return nil
	}

	singleSub1 := GenSubUnaryTreeByLCA(node1, node1, lca)
	singleSub2 := GenSubUnaryTreeByLCA(node2, node2, lca)

	return MergeIntersectSubTrees(singleSub1, singleSub2)
}

// GenLCASubTree 生成树上任意多个节点的最小连通子树
func GenLCASubTree(root *Node, nodeValues []NodeValue) *Node {
	count := len(nodeValues)
	if count == 0 {
		return nil
	} else if count == 1 {
		return &Node{
			Value:    nodeValues[0],
			Parent:   nil,
			Children: nil,
		}
	}

	LinkParent(root)

	var mergedTree *Node = nil
	for i := 0; i < (count - 1); i += 1 {
		subTree := GenTwoNodesLCASubTree(root, nodeValues[i], nodeValues[i+1])
		if subTree == nil {
			return nil
		} else if mergedTree == nil {
			mergedTree = subTree
		} else {
			mergedTree = MergeIntersectSubTrees(subTree, mergedTree)
			if mergedTree == nil {
				return nil
			}
		}
	}

	return mergedTree
}

// GenSubUnaryTreeByLCA 生成从start节点到lca的单叉树
func GenSubUnaryTreeByLCA(start, current, lca *Node) *Node {
	if start.Value.IsEqual(current.Value) {
		current = &Node{
			Value:    start.Value,
			Parent:   start.Parent,
			Children: nil,
		}
	}

	if current.Value.IsEqual(lca.Value) {
		copiedLCA := &Node{
			Value:    lca.Value,
			Parent:   nil,
			Children: current.Children,
		}
		return copiedLCA
	}

	copiedParent := *(current.Parent)
	copiedParent.Children = []*Node{current}
	return GenSubUnaryTreeByLCA(start, &copiedParent, lca)
}
