# Spec 05: Advanced Anomaly Taxonomy & Visualization

## 1. Mathematical Taxonomy

### A5A (Read Skew)
Occurs when transaction $T_1$ reads data item $x$, transaction $T_2$ modifies both $x$ and $y$ and commits, and $T_1$ then reads data item $y$, observing an inconsistent multi-key state:
$$T_1 \xrightarrow{rw} T_2 \xrightarrow{wr} T_1$$

### G0 (Dirty Write)
Occurs when transaction $T_1$ modifies $x$, and before $T_1$ commits or aborts, $T_2$ modifies $x$:
$$T_1 \xrightarrow{ww} T_2 \xrightarrow{ww} T_1$$

### G1c (Circular Information Flow)
Occurs when dependency cycle consists entirely of read-after-write dependencies ($wr$ edges):
$$T_1 \xrightarrow{wr} T_2 \xrightarrow{wr} T_1$$

## 2. Interactive HTML Reporting
The `--export-html` flag emits a single self-contained HTML file containing:
- Embedded CSS styles.
- Vis.js network graph rendering $SG(S) = (V, E)$.
- Chronological worker swimlanes.
- Side-by-side $ddmin$ noise reduction comparison.
