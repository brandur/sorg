$(document).ready ->
  $('a[data-pjax]').pjax
    'timeout': 2000
  hljs.initHighlightingOnLoad()

$(document).on 'pjax:end', ->
  $('pre code').each (i, block) ->
    hljs.highlightBlock(block)
