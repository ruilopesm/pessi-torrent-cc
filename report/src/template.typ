#let project(title: "", authors: (), date: none, logo: none, body) = {
  set document(author: authors.map(a => a.name), title: title)
  set text(font: "New Computer Modern", lang: "pt", region: "pt", size: 11pt)
  show math.equation: set text(weight: 400)

  show par: set block(above: 0.75em, below: 1.5em)

  set par(leading: 0.58em)

  show heading: set block(above: 2em)

  set heading(numbering: "1.")

  show link: underline

  show ref: element => [#lower(element)]

  // Display inline code in a small box
  // that retains the correct baseline.
  show raw.where(block: false): box.with(
    fill: luma(240),
    inset: (x: 3pt, y: 0pt),
    outset: (y: 3pt),
    radius: 2pt,
  )
  
  // Display block code in a larger block
  // with more padding.
  show raw.where(block: true): block.with(
    breakable: false,
    fill: luma(240),
    inset: 10pt,
    radius: 4pt,
  )

  
  // Title page.
  // The page can contain a logo if you pass one with `logo: "logo.png"`.
  if logo != none {
    align(center, image(logo, width: 80%))
  }
  v(9.6fr)

  text(1.1em, date)
  v(1.2em, weak: true)
  text(2em, weight: 700, title)

  pad(
    top: 1em,
    bottom: 0.3em,
    x: 1em,
    grid(
      columns: (1fr,) * 3,
      gutter: 1em,
      ..authors.map(author => align(center)[
        #if author.at("photo", default: none) != none {
          box(stroke: black, image(author.photo))
        }
        *#author.name* \
        #author.number
      ]),
    ),
  )
  
  v(2.4fr)
  pagebreak()


  set par(justify: true)
  set page(numbering: "1", number-align: center)

  body
}