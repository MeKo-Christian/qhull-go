/* Per-step Qhull oracle: dump the facet list (in facet_list order) with each
   facet's vertices and outside set, at the moment Qhull's buildhull has added
   STOPADD vertices (option TA<STOPADD>). This is the pre-pick state for step
   STOPADD: the facet whose outside set's furthest point Qhull picks next.

   Input on stdin: "n stopadd", then n lines "x y".
   Mirrors introspect.c's projection (centroid-subtract) and option string. */
#include "libqhull_r/qhull_ra.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void){
  int n, stopadd;
  if (scanf("%d %d",&n,&stopadd)!=2) return 1;
  double *x=malloc(n*sizeof(double)),*y=malloc(n*sizeof(double));
  for(int i=0;i<n;i++) if(scanf("%lf %lf",&x[i],&y[i])!=2) return 1;
  double xm=0,ym=0; for(int i=0;i<n;i++){xm+=x[i];ym+=y[i];} xm/=n; ym/=n;
  coordT *pts=malloc(n*2*sizeof(coordT));
  for(int i=0;i<n;i++){pts[2*i]=x[i]-xm;pts[2*i+1]=y[i]-ym;}

  char opt[128];
  snprintf(opt,sizeof(opt),"qhull d Qt Qbb Qc Qz TA%d",stopadd);

  qhT qh_qh; qhT *qh=&qh_qh;
  FILE *ef=fopen("/dev/null","w");
  qh_zero(qh,ef);
  int code=qh_new_qhull(qh,2,n,pts,0,opt,NULL,ef);
  (void)code;

  facetT *facet; vertexT *vertex, **vertexp; pointT *point, **pointp;
  printf("VERTICES");
  FORALLvertices printf(" %d:%d", vertex->id, qh_pointid(qh,vertex->point));
  printf("\n");
  printf("FACETLIST (stopadd=%d) facet_next=f%d\n", stopadd, getid_(qh->facet_next));
  FORALLfacets {
    if (facet->visible) continue;
    printf("  f%d up=%d simp=%d verts[", facet->id, facet->upperdelaunay, facet->simplicial);
    FOREACHvertex_(facet->vertices) printf("%d,", qh_pointid(qh,vertex->point));
    printf("] outside[");
    FOREACHpoint_(facet->outsideset) printf("%d,", qh_pointid(qh,point));
    printf("] furthestdist=%.17g\n", facet->furthestdist);
  }
  return 0;
}
