/* Replicate matplotlib's delaunay_impl and dump Qhull internals:
   - global vertex creation order (vertex->id : pointid)
   - each non-upperdelaunay facet's vertex set order, BEFORE qh_triangulate
   Input on stdin: "n", then n lines "x y". */
#include "libqhull_r/qhull_ra.h"
#include <stdio.h>
#include <stdlib.h>

int main(void){
  int n; if (scanf("%d",&n)!=1) return 1;
  double *x=malloc(n*sizeof(double)),*y=malloc(n*sizeof(double));
  for(int i=0;i<n;i++) if(scanf("%lf %lf",&x[i],&y[i])!=2) return 1;
  double xm=0,ym=0; for(int i=0;i<n;i++){xm+=x[i];ym+=y[i];} xm/=n; ym/=n;
  coordT *pts=malloc(n*2*sizeof(coordT));
  for(int i=0;i<n;i++){pts[2*i]=x[i]-xm;pts[2*i+1]=y[i]-ym;}

  qhT qh_qh; qhT *qh=&qh_qh;
  FILE *ef=fopen("/dev/null","w");
  qh_zero(qh,ef);
  int code=qh_new_qhull(qh,2,n,pts,0,(char*)"qhull d Qt Qbb Qc Qz",NULL,ef);
  if(code){printf("ERR %d\n",code);return 1;}

  /* global vertex creation order */
  facetT *facet; vertexT *vertex, **vertexp;
  printf("VERTICES");
  FORALLvertices printf(" %d:%d", vertex->id, qh_pointid(qh,vertex->point));
  printf("\n");

  /* facets BEFORE triangulate */
  printf("FACETS_PRE\n");
  FORALLfacets {
    if (facet->upperdelaunay) continue;
    printf("  f%d toporient=%d simplicial=%d verts", facet->id, facet->toporient, facet->simplicial);
    FOREACHvertex_(facet->vertices) printf(" %d:%d", vertex->id, qh_pointid(qh,vertex->point));
    printf("\n");
  }
  qh_triangulate(qh);
  printf("FACETS_POST\n");
  FORALLfacets {
    if (facet->upperdelaunay) continue;
    int idx[3], k=0;
    FOREACHvertex_(facet->vertices) idx[k++]=qh_pointid(qh,vertex->point);
    printf("  f%d toporient=%d tri %d %d %d\n", facet->id, facet->toporient, idx[0],idx[1],idx[2]);
  }
  return 0;
}
