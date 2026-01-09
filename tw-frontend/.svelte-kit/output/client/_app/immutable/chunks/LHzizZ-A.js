import{x as g,y as u}from"./B9mMlsgT.js";function _(n){const t=n-1;return t*t*t+1}function S(n,{delay:t=0,duration:o=400,easing:s=g}={}){const c=+getComputedStyle(n).opacity;return{delay:t,duration:o,easing:s,css:a=>`opacity: ${a*c}`}}function U(n,{delay:t=0,duration:o=400,easing:s=_,x:c=0,y:a=0,opacity:y=0}={}){const r=getComputedStyle(n),e=+r.opacity,f=r.transform==="none"?"":r.transform,p=e*(1-y),[l,m]=u(c),[$,d]=u(a);return{delay:t,duration:o,easing:s,css:(i,x)=>`
			transform: ${f} translate(${(1-i)*l}${m}, ${(1-i)*$}${d});
			opacity: ${e-p*x}`}}export{S as a,U as f};
