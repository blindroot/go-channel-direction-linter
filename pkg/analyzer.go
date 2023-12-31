package analyzer

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "goFunctionCanDeclareChannelDirection",
	Doc:      "Checks if channel passed as function parameter could have the direction specified",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

type ChannelUsage struct {
	isSentTo   bool
	isRcvdFrom bool

	// Contextual information about the channel
	tokenPos token.Pos
	funcName string
}

func (u *ChannelUsage) isSendOnly() bool {
	return u.isSentTo && !u.isRcvdFrom
}

func (u *ChannelUsage) isReceiveOnly() bool {
	return u.isRcvdFrom && !u.isSentTo
}

func run(pass *analysis.Pass) (interface{}, error) {
	funcInspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	funcInspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}
		channelsWithoutDirection := make(map[string]ChannelUsage)

		params := funcDecl.Type.Params.List
		for _, param := range params {
			chanTypeParam, ok := param.Type.(*ast.ChanType)

			if !ok {
				return
			}

			for _, chanIndent := range param.Names { //case of fun2(ch1,ch2 chan struct{})
				chanName := chanIndent.Name
				if chanTypeParam.Dir != ast.RECV && chanTypeParam.Dir != ast.SEND {
					//fmt.Printf("------> Analysing channel [%s] in function [%s]\n", chanName, funcDecl.Name.Name)
					// if chanName is passed to another func as argument -> ditch the whole analysis
					isPassedAsArgument := false
					ast.Inspect(funcDecl.Body, func(bodyNode ast.Node) bool {
						if callExpr, ok := bodyNode.(*ast.CallExpr); ok {
							for _, arg := range callExpr.Args {
								if name, ok := arg.(*ast.Ident); ok {
									if name.Name == chanName {
										//fmt.Printf("%s used as argument, DITCHING\n", chanName)
										isPassedAsArgument = true
										return false
									}
								}
							}
						}
						return true
					})
					if isPassedAsArgument {
						break
					}
					channelsWithoutDirection[chanName] = ChannelUsage{
						tokenPos: chanIndent.Pos(),
						funcName: funcDecl.Name.Name,
					}

					ast.Inspect(funcDecl.Body, func(bodyNode ast.Node) bool { // search for sending to this channel
						if sendStmt, ok := bodyNode.(*ast.SendStmt); ok {
							if sendStmt.Chan.(*ast.Ident).Name == chanName {
								usage := channelsWithoutDirection[chanName]
								usage.isSentTo = true
								channelsWithoutDirection[chanName] = usage
								return false
							}
						}
						return true
					})
					ast.Inspect(funcDecl.Body, func(bodyNode ast.Node) bool { // search for receiving from this channel
						if recvStmt, ok := bodyNode.(*ast.UnaryExpr); ok {
							if chanIdent, ok := recvStmt.X.(*ast.Ident); ok {
								if recvStmt.Op == token.ARROW && chanIdent.Name == chanName {
									usage := channelsWithoutDirection[chanName]
									usage.isRcvdFrom = true
									channelsWithoutDirection[chanName] = usage
									return false
								}
							}
						}
						return true
					})
				}

			}
		}

		for chName, info := range channelsWithoutDirection {
			if info.isSendOnly() {
				pass.Reportf(info.tokenPos, "Function `%s` uses channel `%s` as send-only, consider `func %s(%s chan<- T`", funcDecl.Name.Name, chName, funcDecl.Name.Name, chName)
			} else if info.isReceiveOnly() {
				pass.Reportf(info.tokenPos, "Function `%s` uses channel `%s` as receive-only, consider `func %s(%s <-chan T`", funcDecl.Name.Name, chName, funcDecl.Name.Name, chName)
			}
		}
		return
	})
	return nil, nil
}
