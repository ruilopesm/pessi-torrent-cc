#let asset_path(name) = "../assets/" + name

#let bdiagbox(
  text_left, text_right,
  width: none, height: none,
  inset: 5pt, text_pad: none,
  box_stroke: none, line_stroke: 1pt,
  inner_width: none,
  left_sep: 0pt, right_sep: 0pt,
  left_outer_sep: 0pt, right_outer_sep: 0pt,
) = style(styles => {
  let left_measure = measure(text_left, styles)
  let right_measure = measure(text_right, styles)

  let text_pad = if text_pad == none {
    // some adjusting; sum 3pt for the base case (5pt)
    // for larger insets, it isn't very relevant
    -2*inset/3 + 3pt
  } else {
    text_pad
  }
  
  let height = if height != none {
    height
  } else {
    2*(left_measure.height + right_measure.height)
  }

  let inner_width = if inner_width != none {
    inner_width
  } else if width != none {
    width - 2*inset
  } else {
    2*calc.max(left_measure.width, right_measure.width)
  }
  
  box(width: inner_width, height: height, stroke: box_stroke)[
    #show line: place.with(top + left)
    #place(top + right, move(dx: -right_sep - text_pad, dy: text_pad, text_right))
    #line(start: (left_outer_sep - inset, -inset), end: (inner_width + inset - right_outer_sep, height + inset), stroke: line_stroke)
    #place(bottom + left, move(dx: left_sep + text_pad, dy: -text_pad, text_left))
  ]
})
#import "@preview/tablex:0.0.6": tablex, rowspanx, colspanx

#let packet_title_color(from, to) = {
  if from == "Node" and to == "Tracker" {
    blue.lighten(50%)
  } else if from == "Tracker" and to == "Node" {
    red.lighten(30%)
  } else if from == "Node" and to == "Node" {
    orange.lighten(30%)
  } else {
    black.lighten(50%)
  }
}

#let packet_direction_box(from, to, background) = box(fill: background, radius: 5pt, inset: 5pt, text(fill: white, size: 10pt, font: "Fira Sans", [#from $->$ #to]))

//#let packet_title(title, from, to) = heading(numbering: none, level: 2)[#title #box(pad(bottom: -4pt, packet_direction_box(from, to, packet_title_color(from, to))))]

#let packet_title(title, from, to) = [#place( left, packet_direction_box(from, to, packet_title_color(from, to))) *#title*]

#let packet_def(packet_id, packet_name, from, to, ..c) = block(breakable: false, tablex(
  columns: (25%, 20%, 1fr),
  rows: (25pt, 15pt, auto),
  align: center + horizon,
  stroke: 0.5pt + gray.darken(70%),
  fill: (col, row) => {
    if row == 1 {
      gray.lighten(30%)
    } else if col == 0 and row > 1 {
      gray.lighten(50%)
    } else {
      white
    }
  },
  colspanx(fill: gray.lighten(20%), 3)[#packet_title(packet_name, from, to)],
  [Campo], [Tipo de dados], [Descrição],
  [PacketType], [u8], [Sempre igual a #packet_id],
  ..c
))
