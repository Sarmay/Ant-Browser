import {
  useState,
  useRef,
  useEffect,
  useCallback,
  useMemo,
  type ReactNode,
  type KeyboardEvent,
  type UIEvent,
} from 'react'
import clsx from 'clsx'
import {
  CheckCircle,
  Copy,
  Maximize2,
  Minimize2,
  Sparkles,
  Code2,
} from 'lucide-react'
import { Button, toast } from '../components'

export type CodeLanguage = 'javascript' | 'json' | 'typescript' | 'text'

export type CodeTokenKind =
  | 'plain'
  | 'comment'
  | 'string'
  | 'template-string'
  | 'template-expr'
  | 'keyword'
  | 'builtin'
  | 'function'
  | 'number'
  | 'boolean'
  | 'null'
  | 'regex'
  | 'property'
  | 'operator'
  | 'punctuation'

export interface CodeToken {
  content: string
  kind: CodeTokenKind
}

interface ParserLineState {
  inBlockComment: boolean
  inTemplateString: boolean
  templateBraceStack: number[]
}

const JS_KEYWORDS = new Set([
  'async',
  'await',
  'break',
  'case',
  'catch',
  'class',
  'const',
  'continue',
  'debugger',
  'default',
  'delete',
  'do',
  'else',
  'export',
  'extends',
  'finally',
  'for',
  'from',
  'function',
  'get',
  'if',
  'import',
  'in',
  'instanceof',
  'let',
  'new',
  'of',
  'return',
  'set',
  'static',
  'super',
  'switch',
  'this',
  'throw',
  'try',
  'typeof',
  'var',
  'void',
  'while',
  'with',
  'yield',
  'as',
])

const JS_BUILTINS = new Set([
  'Array',
  'Boolean',
  'Buffer',
  'Date',
  'Error',
  'JSON',
  'Map',
  'Math',
  'Number',
  'Object',
  'Promise',
  'RegExp',
  'Set',
  'String',
  'Symbol',
  'WeakMap',
  'WeakSet',
  'browser',
  'clearInterval',
  'clearTimeout',
  'console',
  'context',
  'document',
  'exports',
  'fetch',
  'global',
  'globalThis',
  'module',
  'page',
  'params',
  'playwright',
  'process',
  'puppeteer',
  'require',
  'response',
  'selector',
  'setInterval',
  'setTimeout',
  'window',
])

const JS_CONSTANTS = new Set([
  'true',
  'false',
  'null',
  'undefined',
  'NaN',
  'Infinity',
])

const TOKEN_CLASS_NAMES: Record<CodeTokenKind, string> = {
  plain: 'text-[var(--color-text-primary)]',
  comment: 'text-[#8b949e] dark:text-[#8b949e] italic opacity-90',
  string: 'text-[#0a8043] dark:text-[#7ee787]',
  'template-string': 'text-[#0a8043] dark:text-[#7ee787]',
  'template-expr': 'text-[#005cc5] dark:text-[#79c0ff] font-semibold',
  keyword: 'text-[#cf222e] dark:text-[#ff7b72] font-medium',
  builtin: 'text-[#8250df] dark:text-[#d2a8ff] font-medium',
  function: 'text-[#0969da] dark:text-[#79c0ff]',
  number: 'text-[#953800] dark:text-[#ffa657]',
  boolean: 'text-[#0550ae] dark:text-[#79c0ff] font-medium',
  null: 'text-[#0550ae] dark:text-[#79c0ff] font-medium',
  regex: 'text-[#032f62] dark:text-[#a5d6ff]',
  property: 'text-[#116329] dark:text-[#93c5fd]',
  operator: 'text-[#0550ae] dark:text-[#ff7b72]',
  punctuation: 'text-[#57606a] dark:text-[#c9d1d9]',
}

function tokenizeJsLine(
  line: string,
  state: ParserLineState,
): { tokens: CodeToken[]; nextState: ParserLineState } {
  const tokens: CodeToken[] = []
  let i = 0
  const len = line.length
  let inBlockComment = state.inBlockComment
  let inTemplateString = state.inTemplateString
  const templateBraceStack = [...state.templateBraceStack]

  let prevNonWhitespaceKind: CodeTokenKind | null = null

  while (i < len) {
    if (inBlockComment) {
      const endCommentIndex = line.indexOf('*/', i)
      if (endCommentIndex === -1) {
        tokens.push({ content: line.slice(i), kind: 'comment' })
        i = len
      } else {
        tokens.push({
          content: line.slice(i, endCommentIndex + 2),
          kind: 'comment',
        })
        i = endCommentIndex + 2
        inBlockComment = false
        prevNonWhitespaceKind = 'comment'
      }
      continue
    }

    if (inTemplateString) {
      let chunk = ''
      let escaped = false
      let closed = false
      let exprStart = false

      while (i < len) {
        const char = line[i]
        if (escaped) {
          chunk += char
          escaped = false
          i++
        } else if (char === '\\') {
          chunk += char
          escaped = true
          i++
        } else if (char === '`') {
          chunk += char
          i++
          closed = true
          break
        } else if (char === '$' && i + 1 < len && line[i + 1] === '{') {
          exprStart = true
          break
        } else {
          chunk += char
          i++
        }
      }

      if (chunk) {
        tokens.push({ content: chunk, kind: 'template-string' })
        prevNonWhitespaceKind = 'string'
      }

      if (closed) {
        inTemplateString = false
      } else if (exprStart) {
        tokens.push({ content: '${', kind: 'template-expr' })
        i += 2
        inTemplateString = false
        templateBraceStack.push(1)
        prevNonWhitespaceKind = 'punctuation'
      }
      continue
    }

    const char = line[i]

    if (char === ' ' || char === '\t' || char === '\r') {
      let space = ''
      while (
        i < len &&
        (line[i] === ' ' || line[i] === '\t' || line[i] === '\r')
      ) {
        space += line[i]
        i++
      }
      tokens.push({ content: space, kind: 'plain' })
      continue
    }

    if (char === '/' && i + 1 < len && line[i + 1] === '/') {
      tokens.push({ content: line.slice(i), kind: 'comment' })
      i = len
      prevNonWhitespaceKind = 'comment'
      break
    }

    if (char === '/' && i + 1 < len && line[i + 1] === '*') {
      const endCommentIndex = line.indexOf('*/', i + 2)
      if (endCommentIndex === -1) {
        tokens.push({ content: line.slice(i), kind: 'comment' })
        inBlockComment = true
        i = len
      } else {
        tokens.push({
          content: line.slice(i, endCommentIndex + 2),
          kind: 'comment',
        })
        i = endCommentIndex + 2
      }
      prevNonWhitespaceKind = 'comment'
      continue
    }

    if (char === "'" || char === '"') {
      const quote = char
      let str = quote
      i++
      let escaped = false
      while (i < len) {
        const c = line[i]
        str += c
        i++
        if (escaped) {
          escaped = false
        } else if (c === '\\') {
          escaped = true
        } else if (c === quote) {
          break
        }
      }
      tokens.push({ content: str, kind: 'string' })
      prevNonWhitespaceKind = 'string'
      continue
    }

    if (char === '`') {
      let str = '`'
      i++
      let escaped = false
      let exprStart = false
      let closed = false
      while (i < len) {
        const c = line[i]
        if (escaped) {
          str += c
          escaped = false
          i++
        } else if (c === '\\') {
          str += c
          escaped = true
          i++
        } else if (c === '`') {
          str += c
          i++
          closed = true
          break
        } else if (c === '$' && i + 1 < len && line[i + 1] === '{') {
          exprStart = true
          break
        } else {
          str += c
          i++
        }
      }
      tokens.push({ content: str, kind: 'template-string' })
      prevNonWhitespaceKind = 'string'
      if (closed) {
        inTemplateString = false
      } else if (exprStart) {
        tokens.push({ content: '${', kind: 'template-expr' })
        i += 2
        inTemplateString = false
        templateBraceStack.push(1)
        prevNonWhitespaceKind = 'punctuation'
      } else {
        inTemplateString = true
      }
      continue
    }

    if (
      char === '/' &&
      (prevNonWhitespaceKind === null ||
        prevNonWhitespaceKind === 'operator' ||
        prevNonWhitespaceKind === 'punctuation' ||
        prevNonWhitespaceKind === 'keyword')
    ) {
      let regexStr = '/'
      i++
      let escaped = false
      let inCharClass = false
      let closed = false
      while (i < len) {
        const c = line[i]
        regexStr += c
        i++
        if (escaped) {
          escaped = false
        } else if (c === '\\') {
          escaped = true
        } else if (c === '[' && !inCharClass) {
          inCharClass = true
        } else if (c === ']' && inCharClass) {
          inCharClass = false
        } else if (c === '/' && !inCharClass) {
          closed = true
          break
        }
      }
      if (closed) {
        while (i < len && /[a-z]/i.test(line[i])) {
          regexStr += line[i]
          i++
        }
        tokens.push({ content: regexStr, kind: 'regex' })
        prevNonWhitespaceKind = 'regex'
        continue
      }
    }

    if (
      /\d/.test(char) ||
      (char === '.' && i + 1 < len && /\d/.test(line[i + 1]))
    ) {
      let num = ''
      if (char === '0' && i + 1 < len && (line[i + 1] === 'x' || line[i + 1] === 'X')) {
        num += line.slice(i, i + 2)
        i += 2
        while (i < len && /[0-9a-fA-F_]/.test(line[i])) {
          num += line[i]
          i++
        }
      } else if (char === '0' && i + 1 < len && (line[i + 1] === 'b' || line[i + 1] === 'B')) {
        num += line.slice(i, i + 2)
        i += 2
        while (i < len && /[01_]/.test(line[i])) {
          num += line[i]
          i++
        }
      } else {
        while (i < len && /[\d._eE+-]/.test(line[i])) {
          if (
            (line[i] === '+' || line[i] === '-') &&
            i > 0 &&
            !/[eE]/.test(line[i - 1])
          ) {
            break
          }
          num += line[i]
          i++
        }
        if (i < len && line[i] === 'n') {
          num += 'n'
          i++
        }
      }
      tokens.push({ content: num, kind: 'number' })
      prevNonWhitespaceKind = 'number'
      continue
    }

    if (/[a-zA-Z_$]/.test(char)) {
      let word = ''
      while (i < len && /[a-zA-Z0-9_$]/.test(line[i])) {
        word += line[i]
        i++
      }

      let kind: CodeTokenKind = 'plain'
      if (JS_KEYWORDS.has(word)) {
        kind = 'keyword'
      } else if (JS_CONSTANTS.has(word)) {
        kind = word === 'null' || word === 'undefined' ? 'null' : 'boolean'
      } else if (JS_BUILTINS.has(word)) {
        kind = 'builtin'
      } else {
        let peek = i
        while (peek < len && (line[peek] === ' ' || line[peek] === '\t')) {
          peek++
        }
        if (peek < len && line[peek] === '(') {
          kind = 'function'
        } else if (prevNonWhitespaceKind === 'punctuation' && tokens.some(t => t.content === '.')) {
          kind = 'property'
        }
      }

      tokens.push({ content: word, kind })
      prevNonWhitespaceKind = kind
      continue
    }

    if (templateBraceStack.length > 0) {
      if (char === '{') {
        templateBraceStack[templateBraceStack.length - 1]++
      } else if (char === '}') {
        templateBraceStack[templateBraceStack.length - 1]--
        if (templateBraceStack[templateBraceStack.length - 1] === 0) {
          templateBraceStack.pop()
          tokens.push({ content: '}', kind: 'template-expr' })
          i++
          inTemplateString = true
          prevNonWhitespaceKind = 'punctuation'
          continue
        }
      }
    }

    const op3 = line.slice(i, i + 3)
    if (op3 === '===' || op3 === '!==' || op3 === '...' || op3 === '>>>' || op3 === '??=') {
      tokens.push({ content: op3, kind: 'operator' })
      i += 3
      prevNonWhitespaceKind = 'operator'
      continue
    }

    const op2 = line.slice(i, i + 2)
    if (
      [
        '==',
        '!=',
        '<=',
        '>=',
        '&&',
        '||',
        '??',
        '?.',
        '=>',
        '++',
        '--',
        '+=',
        '-=',
        '*=',
        '/=',
        '%=',
        '&=',
        '|=',
        '^=',
        '<<',
        '>>',
      ].includes(op2)
    ) {
      tokens.push({ content: op2, kind: 'operator' })
      i += 2
      prevNonWhitespaceKind = 'operator'
      continue
    }

    if (['+', '-', '*', '/', '%', '=', '<', '>', '!', '?', ':', '&', '|', '^', '~'].includes(char)) {
      tokens.push({ content: char, kind: 'operator' })
      i++
      prevNonWhitespaceKind = 'operator'
      continue
    }

    tokens.push({ content: char, kind: 'punctuation' })
    i++
    prevNonWhitespaceKind = 'punctuation'
  }

  return {
    tokens: tokens.length ? tokens : [{ content: '', kind: 'plain' }],
    nextState: {
      inBlockComment,
      inTemplateString,
      templateBraceStack,
    },
  }
}

const JSON_PATTERN =
  /("(?:\\.|[^"\\])*")(?=\s*:)|("(?:\\.|[^"\\])*")|\b(?:true|false)\b|\bnull\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?|[{}\[\],:]/g

function tokenizeJsonLine(line: string): CodeToken[] {
  const tokens: CodeToken[] = []
  let lastIndex = 0
  JSON_PATTERN.lastIndex = 0

  for (let match = JSON_PATTERN.exec(line); match; match = JSON_PATTERN.exec(line)) {
    if (match.index > lastIndex) {
      tokens.push({ content: line.slice(lastIndex, match.index), kind: 'plain' })
    }
    const [token] = match
    let kind: CodeTokenKind = 'punctuation'
    if (match[1]) {
      kind = 'property'
    } else if (match[2]) {
      kind = 'string'
    } else if (token === 'true' || token === 'false') {
      kind = 'boolean'
    } else if (token === 'null') {
      kind = 'null'
    } else if (/^-?\d/.test(token)) {
      kind = 'number'
    }
    tokens.push({ content: token, kind })
    lastIndex = match.index + token.length
  }

  if (lastIndex < line.length) {
    tokens.push({ content: line.slice(lastIndex), kind: 'plain' })
  }

  return tokens.length ? tokens : [{ content: '', kind: 'plain' }]
}

export function formatJsCode(code: string): string {
  const lines = code.split('\n')
  let indentLevel = 0
  const result: string[] = []

  for (let i = 0; i < lines.length; i++) {
    const rawLine = lines[i]
    const trimmed = rawLine.trim()

    if (!trimmed) {
      result.push('')
      continue
    }

    let dedentCount = 0
    let temp = trimmed
    while (
      temp.startsWith('}') ||
      temp.startsWith(']') ||
      temp.startsWith(')')
    ) {
      dedentCount++
      temp = temp.slice(1).trim()
    }

    const currentIndent = Math.max(0, indentLevel - dedentCount)
    const formattedLine = '  '.repeat(currentIndent) + trimmed
    result.push(formattedLine)

    for (let j = 0; j < trimmed.length; j++) {
      const char = trimmed[j]
      if (char === '{' || char === '[' || char === '(') {
        indentLevel++
      } else if (char === '}' || char === ']' || char === ')') {
        indentLevel = Math.max(0, indentLevel - 1)
      }
    }
  }

  return result.join('\n')
}

export interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  language?: CodeLanguage
  readOnly?: boolean
  disabled?: boolean
  placeholder?: string
  minHeight?: string | number
  maxHeight?: string | number
  className?: string
  showLineNumbers?: boolean
  showToolbar?: boolean
  title?: string
  autoFocus?: boolean
}

export function CodeEditor({
  value,
  onChange,
  language = 'javascript',
  readOnly = false,
  disabled = false,
  placeholder,
  minHeight = '480px',
  maxHeight,
  className,
  showLineNumbers = true,
  showToolbar = true,
  title,
  autoFocus = false,
}: CodeEditorProps) {
  const [copied, setCopied] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [cursorPos, setCursorPos] = useState({ line: 1, col: 1 })

  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const preRef = useRef<HTMLPreElement>(null)
  const gutterRef = useRef<HTMLDivElement>(null)

  const isJs = language === 'javascript' || language === 'typescript'
  const isJson = language === 'json'

  const lines = useMemo(() => {
    return value.split('\n')
  }, [value])

  const lineCount = lines.length
  const charCount = value.length

  const tokenizedLines = useMemo(() => {
    if (isJson) {
      return lines.map((line) => tokenizeJsonLine(line))
    }
    if (!isJs) {
      return lines.map((line) => [{ content: line, kind: 'plain' as const }])
    }

    let state: ParserLineState = {
      inBlockComment: false,
      inTemplateString: false,
      templateBraceStack: [],
    }

    return lines.map((line) => {
      const { tokens, nextState } = tokenizeJsLine(line, state)
      state = nextState
      return tokens
    })
  }, [lines, isJs, isJson])

  const updateCursorPosition = useCallback(() => {
    if (!textareaRef.current) return
    const selStart = textareaRef.current.selectionStart
    const textBefore = value.slice(0, selStart)
    const linesBefore = textBefore.split('\n')
    const currentLine = linesBefore.length
    const currentCol = linesBefore[linesBefore.length - 1].length + 1
    setCursorPos({ line: currentLine, col: currentCol })
  }, [value])

  const handleScroll = (e: UIEvent<HTMLTextAreaElement>) => {
    const target = e.currentTarget
    if (preRef.current) {
      preRef.current.scrollTop = target.scrollTop
      preRef.current.scrollLeft = target.scrollLeft
    }
    if (gutterRef.current) {
      gutterRef.current.scrollTop = target.scrollTop
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (readOnly || disabled) return
    const textarea = e.currentTarget
    const { selectionStart, selectionEnd, value: currentVal } = textarea

    if (e.key === 'Tab') {
      e.preventDefault()
      if (e.shiftKey) {
        const lineStart = currentVal.lastIndexOf('\n', selectionStart - 1) + 1
        const lineEnd = currentVal.indexOf('\n', selectionEnd)
        const actualEnd = lineEnd === -1 ? currentVal.length : lineEnd
        const selectedBlock = currentVal.slice(lineStart, actualEnd)
        const blockLines = selectedBlock.split('\n')
        const modified = blockLines
          .map((l) => (l.startsWith('  ') ? l.slice(2) : l.startsWith(' ') ? l.slice(1) : l))
          .join('\n')
        const nextVal =
          currentVal.slice(0, lineStart) + modified + currentVal.slice(actualEnd)
        onChange?.(nextVal)
        setTimeout(() => {
          if (textareaRef.current) {
            textareaRef.current.selectionStart = Math.max(
              lineStart,
              selectionStart - 2,
            )
            textareaRef.current.selectionEnd =
              selectionStart === selectionEnd
                ? Math.max(lineStart, selectionStart - 2)
                : lineStart + modified.length
            updateCursorPosition()
          }
        }, 0)
      } else {
        if (selectionStart === selectionEnd) {
          const nextVal =
            currentVal.slice(0, selectionStart) +
            '  ' +
            currentVal.slice(selectionEnd)
          onChange?.(nextVal)
          setTimeout(() => {
            if (textareaRef.current) {
              textareaRef.current.selectionStart =
                textareaRef.current.selectionEnd = selectionStart + 2
              updateCursorPosition()
            }
          }, 0)
        } else {
          const lineStart = currentVal.lastIndexOf('\n', selectionStart - 1) + 1
          const lineEnd = currentVal.indexOf('\n', selectionEnd)
          const actualEnd = lineEnd === -1 ? currentVal.length : lineEnd
          const selectedBlock = currentVal.slice(lineStart, actualEnd)
          const blockLines = selectedBlock.split('\n')
          const modified = blockLines.map((l) => '  ' + l).join('\n')
          const nextVal =
            currentVal.slice(0, lineStart) + modified + currentVal.slice(actualEnd)
          onChange?.(nextVal)
          setTimeout(() => {
            if (textareaRef.current) {
              textareaRef.current.selectionStart = lineStart
              textareaRef.current.selectionEnd = lineStart + modified.length
              updateCursorPosition()
            }
          }, 0)
        }
      }
      return
    }

    if (e.key === 'Enter') {
      const lineStart = currentVal.lastIndexOf('\n', selectionStart - 1) + 1
      const currentLineText = currentVal.slice(lineStart, selectionStart)
      const indentMatch = currentLineText.match(/^(\s*)/)
      let indent = indentMatch ? indentMatch[1] : ''

      const prevChar = currentVal[selectionStart - 1]
      const nextChar = currentVal[selectionStart]

      if (
        (prevChar === '{' && nextChar === '}') ||
        (prevChar === '[' && nextChar === ']') ||
        (prevChar === '(' && nextChar === ')')
      ) {
        e.preventDefault()
        const insertText = '\n' + indent + '  \n' + indent
        const nextVal =
          currentVal.slice(0, selectionStart) +
          insertText +
          currentVal.slice(selectionEnd)
        onChange?.(nextVal)
        setTimeout(() => {
          if (textareaRef.current) {
            textareaRef.current.selectionStart =
              textareaRef.current.selectionEnd =
                selectionStart + indent.length + 3
            updateCursorPosition()
          }
        }, 0)
        return
      }

      if (prevChar === '{' || prevChar === '[' || prevChar === '(') {
        indent += '  '
      }

      if (indent) {
        e.preventDefault()
        const insertText = '\n' + indent
        const nextVal =
          currentVal.slice(0, selectionStart) +
          insertText +
          currentVal.slice(selectionEnd)
        onChange?.(nextVal)
        setTimeout(() => {
          if (textareaRef.current) {
            textareaRef.current.selectionStart =
              textareaRef.current.selectionEnd = selectionStart + insertText.length
            updateCursorPosition()
          }
        }, 0)
        return
      }
    }

    const pairMap: Record<string, string> = {
      '(': ')',
      '[': ']',
      '{': '}',
      '"': '"',
      "'": "'",
      '`': '`',
    }

    if (pairMap[e.key] && selectionStart === selectionEnd) {
      const nextChar = currentVal[selectionStart]
      if (
        e.key === nextChar &&
        (e.key === '"' ||
          e.key === "'" ||
          e.key === '`' ||
          e.key === ')' ||
          e.key === ']' ||
          e.key === '}')
      ) {
        e.preventDefault()
        textarea.selectionStart = textarea.selectionEnd = selectionStart + 1
        updateCursorPosition()
        return
      }

      e.preventDefault()
      const closingChar = pairMap[e.key]
      const nextVal =
        currentVal.slice(0, selectionStart) +
        e.key +
        closingChar +
        currentVal.slice(selectionEnd)
      onChange?.(nextVal)
      setTimeout(() => {
        if (textareaRef.current) {
          textareaRef.current.selectionStart =
            textareaRef.current.selectionEnd = selectionStart + 1
          updateCursorPosition()
        }
      }, 0)
      return
    }

    if (
      e.key === 'Backspace' &&
      selectionStart === selectionEnd &&
      selectionStart > 0
    ) {
      const prevC = currentVal[selectionStart - 1]
      const nextC = currentVal[selectionStart]
      if (
        (prevC === '(' && nextC === ')') ||
        (prevC === '[' && nextC === ']') ||
        (prevC === '{' && nextC === '}') ||
        (prevC === '"' && nextC === '"') ||
        (prevC === "'" && nextC === "'") ||
        (prevC === '`' && nextC === '`')
      ) {
        e.preventDefault()
        const nextVal =
          currentVal.slice(0, selectionStart - 1) +
          currentVal.slice(selectionStart + 1)
        onChange?.(nextVal)
        setTimeout(() => {
          if (textareaRef.current) {
            textareaRef.current.selectionStart =
              textareaRef.current.selectionEnd = selectionStart - 1
            updateCursorPosition()
          }
        }, 0)
        return
      }
    }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      toast.success('代码已复制')
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('复制失败')
    }
  }

  const handleFormat = () => {
    if (readOnly || disabled || !onChange) return
    try {
      if (isJson) {
        const parsed = JSON.parse(value)
        onChange(JSON.stringify(parsed, null, 2))
        toast.success('JSON 已格式化')
      } else if (isJs) {
        const formatted = formatJsCode(value)
        onChange(formatted)
        toast.success('代码缩进已格式化')
      }
    } catch (e) {
      toast.warning('代码存在语法错误，无法自动格式化')
    }
  }

  useEffect(() => {
    if (autoFocus && textareaRef.current) {
      textareaRef.current.focus()
    }
  }, [autoFocus])

  const renderHighlightedContent = (): ReactNode => {
    return tokenizedLines.map((lineTokens, lineIdx) => {
      const isCurrentLine = cursorPos.line === lineIdx + 1
      return (
        <div
          key={`line-${lineIdx}`}
          className={clsx(
            'flex leading-[20px] transition-colors',
            isCurrentLine && !readOnly && 'bg-blue-500/[0.04] dark:bg-blue-400/[0.06]',
          )}
        >
          {lineTokens.map((tok, tokIdx) => (
            <span
              key={`tok-${lineIdx}-${tokIdx}`}
              className={TOKEN_CLASS_NAMES[tok.kind]}
            >
              {tok.content || '\u00a0'}
            </span>
          ))}
          {lineTokens.length === 0 || (lineTokens.length === 1 && !lineTokens[0].content) ? (
            <span>{'\u00a0'}</span>
          ) : null}
        </div>
      )
    })
  }

  const editorCore = (
    <div
      className={clsx(
        'relative flex w-full overflow-hidden rounded-b-xl bg-[var(--color-bg-surface)] text-left transition-colors',
        disabled && 'opacity-60 cursor-not-allowed',
        className,
      )}
      style={{
        minHeight: isFullscreen ? 'calc(100vh - 120px)' : minHeight,
        maxHeight: isFullscreen ? 'calc(100vh - 120px)' : maxHeight,
      }}
    >
      {showLineNumbers && (
        <div
          ref={gutterRef}
          aria-hidden="true"
          className="select-none overflow-hidden border-r border-[var(--color-border-muted)] bg-[var(--color-bg-secondary)]/50 py-3 pl-2 pr-3 text-right font-mono text-xs leading-[20px] text-[var(--color-text-muted)] opacity-70"
          style={{ width: `${Math.max(40, String(lineCount).length * 9 + 20)}px` }}
        >
          {Array.from({ length: lineCount }).map((_, idx) => {
            const lineNum = idx + 1
            const isActive = cursorPos.line === lineNum
            return (
              <div
                key={`gutter-${lineNum}`}
                className={clsx(
                  'transition-colors',
                  isActive && !readOnly && 'font-bold text-[var(--color-accent)] opacity-100',
                )}
              >
                {lineNum}
              </div>
            )
          })}
        </div>
      )}

      <div className="relative min-w-0 flex-1 overflow-hidden">
        <pre
          ref={preRef}
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 z-0 m-0 overflow-hidden p-3 font-mono text-xs leading-[20px] whitespace-pre select-none"
          style={{
            tabSize: 2,
            fontFamily:
              'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
          }}
        >
          <code className="block min-w-full">{renderHighlightedContent()}</code>
        </pre>

        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => {
            onChange?.(e.target.value)
            updateCursorPosition()
          }}
          onKeyDown={handleKeyDown}
          onKeyUp={updateCursorPosition}
          onClick={updateCursorPosition}
          onSelect={updateCursorPosition}
          onScroll={handleScroll}
          disabled={disabled}
          readOnly={readOnly}
          spellCheck={false}
          autoCapitalize="off"
          autoComplete="off"
          autoCorrect="off"
          placeholder={placeholder}
          className={clsx(
            'relative z-10 block h-full w-full resize-none overflow-auto border-none bg-transparent p-3 font-mono text-xs leading-[20px] whitespace-pre outline-none',
            'selection:bg-blue-500/25 dark:selection:bg-blue-400/30 text-transparent caret-[var(--color-text-primary)]',
            disabled && 'cursor-not-allowed',
          )}
          style={{
            tabSize: 2,
            fontFamily:
              'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
          }}
        />
      </div>
    </div>
  )

  const languageLabel =
    language === 'javascript'
      ? 'JavaScript (Node.js)'
      : language === 'typescript'
        ? 'TypeScript'
        : language === 'json'
          ? 'JSON'
          : 'Plain Text'

  const toolbar = showToolbar && (
    <div className="flex flex-wrap items-center justify-between gap-2 border-b border-[var(--color-border-muted)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)]">
      <div className="flex items-center gap-2">
        <span className="inline-flex items-center gap-1 rounded-md border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-2 py-0.5 text-[11px] font-semibold text-[var(--color-text-primary)]">
          <Code2 className="h-3 w-3 text-[var(--color-accent)]" />
          {title || languageLabel}
        </span>
        <span className="text-[var(--color-text-muted)]">
          {lineCount} 行 · {charCount} 字符
        </span>
        {!readOnly && (
          <span className="hidden sm:inline text-[var(--color-text-muted)]">
            · 行 {cursorPos.line}, 列 {cursorPos.col}
          </span>
        )}
      </div>

      <div className="flex items-center gap-1.5">
        {!readOnly && (
          <Button
            type="button"
            variant="secondary"
            size="sm"
            className="!h-7 !px-2 text-xs"
            onClick={handleFormat}
            disabled={disabled || !value.trim()}
            title="格式化缩进"
          >
            <Sparkles className="h-3.5 w-3.5" />
            格式化
          </Button>
        )}
        <Button
          type="button"
          variant="secondary"
          size="sm"
          className="!h-7 !px-2 text-xs"
          onClick={() => void handleCopy()}
          title="复制全部代码"
        >
          {copied ? (
            <CheckCircle className="h-3.5 w-3.5 text-green-500" />
          ) : (
            <Copy className="h-3.5 w-3.5" />
          )}
          {copied ? '已复制' : '复制'}
        </Button>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          className="!h-7 !px-2 text-xs"
          onClick={() => setIsFullscreen((prev) => !prev)}
          title={isFullscreen ? '退出全屏' : '全屏编辑'}
        >
          {isFullscreen ? (
            <Minimize2 className="h-3.5 w-3.5" />
          ) : (
            <Maximize2 className="h-3.5 w-3.5" />
          )}
          {isFullscreen ? '还原' : '全屏'}
        </Button>
      </div>
    </div>
  )

  if (isFullscreen) {
    return (
      <div className="fixed inset-0 z-50 flex flex-col bg-[var(--color-bg-base)] p-4 animate-fade-in">
        <div className="mb-3 flex items-center justify-between rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-2.5 shadow-[var(--shadow-sm)]">
          <div className="flex items-center gap-2 text-sm font-semibold text-[var(--color-text-primary)]">
            <Code2 className="h-4 w-4 text-[var(--color-accent)]" />
            <span>{title || '代码编辑器'}</span>
            <span className="rounded bg-[var(--color-bg-muted)] px-1.5 py-0.5 text-xs font-normal text-[var(--color-text-muted)]">
              {languageLabel}
            </span>
          </div>
          <div className="flex items-center gap-2">
            {!readOnly && (
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={handleFormat}
                disabled={disabled}
              >
                <Sparkles className="h-3.5 w-3.5" />
                格式化
              </Button>
            )}
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => void handleCopy()}
            >
              {copied ? (
                <CheckCircle className="h-3.5 w-3.5 text-green-500" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
              {copied ? '已复制' : '复制'}
            </Button>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => setIsFullscreen(false)}
            >
              <Minimize2 className="h-3.5 w-3.5" />
              退出全屏
            </Button>
          </div>
        </div>

        <div className="flex-1 overflow-hidden rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] shadow-[var(--shadow-md)] flex flex-col">
          {toolbar}
          <div className="flex-1 overflow-hidden">{editorCore}</div>
        </div>
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] shadow-[var(--shadow-sm)] transition-colors focus-within:border-[var(--color-border-strong)] focus-within:ring-1 focus-within:ring-[var(--color-border-strong)]">
      {toolbar}
      {editorCore}
    </div>
  )
}
