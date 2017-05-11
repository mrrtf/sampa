# Sampa

## Decodes Alice MCH raw data coming out from DualSampa + G-RORC setup

+++

# Written in Go

> warning: that's my first Go package **ever**

---

# Readout is done with DATE still

---

# Data structure

Is described in [Filippo's gitlab](https://gitlab.cern.ch/costaf/grorc)
(under DATE DATA FORMAT (GBT MODE))

+++

In the DATE payload, each GBT word (80bits) is coded in 4 DATE words

+++

words | comment
--------------------------------- | --------------------
[32bits] x 7 | 7 words added by the software (fake equipment header) 
00000000 00000000 00000000 00000001 | SOP (required)
0x0[127:96] 0x0[95:64] GBT[63:32] GBT[31:0] | first GBT word
0x0[127:96] 0x0[95:64] GBT[63:32] GBT[31:0] | second GBT word
... | ... 
00000000 00000000 00000000 0000XXXX | EOP (optional)

