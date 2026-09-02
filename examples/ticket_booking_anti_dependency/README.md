# Ticket Seat Reservation Anti-Dependency Cycle ($G2$)

## Business Context
In airline and concert ticket reservation platforms, booking algorithms attempt to keep adjacent seats together or reserve buffer rows. Three concurrent users attempt to book seats:
- User 1 checks seat 2, then books seat 1.
- User 2 checks seat 3, then books seat 2.
- User 3 checks seat 1, then books seat 3.

## Mathematical Formulation ($G2$)
According to Atul Adya (1999) and Berenson et al. (1995), a generalized $G2$ Anti-Dependency cycle occurs when a directed serialization graph $SG(S)$ contains a cycle consisting purely of anti-dependency edges ($\xrightarrow{rw}$) across $k \ge 3$ transactions:

$$T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_3 \xrightarrow{rw} T_1$$

## Remediation
Under `SERIALIZABLE` isolation or Strict Two-Phase Locking (S2PL), the database detects the anti-dependency cycle and aborts one of the conflicting transactions with a serialization failure (`SQLSTATE 40001`).
